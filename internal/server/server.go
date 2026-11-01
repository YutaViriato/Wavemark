// Package server is the Wavemark HTTP runtime: it serves the current site
// snapshot, rebuilds it when files change and pushes reloads in dev mode.
package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/site"
	"github.com/YutaViriato/Wavemark/internal/theme"
	"github.com/fsnotify/fsnotify"
)

// Server serves one content folder.
type Server struct {
	root  string
	dev   bool
	theme *theme.Theme
	log   *log.Logger

	site     atomic.Pointer[site.Site]
	notFound atomic.Pointer[[]byte]
	fontsCSS atomic.Pointer[[]byte]

	buildMu sync.Mutex
	subsMu  sync.Mutex
	subs    map[chan struct{}]struct{}
	done    chan struct{} // closed on shutdown so event streams end
	bootID  string        // changes on every start; dev clients reload when it changes
}

// New builds the first snapshot. In serve mode (dev=false) it fails when the
// content has errors, so a broken site never goes live.
func New(root string, dev bool, logger *log.Logger) (*Server, error) {
	th, err := theme.New()
	if err != nil {
		return nil, err
	}
	s := &Server{root: root, dev: dev, theme: th, log: logger, subs: map[chan struct{}]struct{}{}, done: make(chan struct{}), bootID: strconv.FormatInt(time.Now().UnixNano(), 36)}
	st, err := s.build(nil)
	if err != nil {
		return nil, err
	}
	if !dev && diag.HasErrors(st.Diags) {
		return nil, &ContentError{Diags: st.Diags}
	}
	s.swap(st)
	return s, nil
}

// ContentError reports diagnostics that prevent serving.
type ContentError struct{ Diags []diag.Diagnostic }

func (e *ContentError) Error() string {
	n, _ := diag.Count(e.Diags)
	return fmt.Sprintf("the content has %d error(s); fix them with 'wavemark check' before serving", n)
}

func (s *Server) build(prev *site.Site) (*site.Site, error) {
	opts := theme.Options{Dev: s.dev}
	return site.Build(site.Options{
		Root: s.root,
		Dev:  s.dev,
		Prev: prev,
		Render: func(st *site.Site, p *site.Page) ([]byte, error) {
			return s.theme.RenderPage(st, p, opts)
		},
	})
}

func (s *Server) swap(st *site.Site) {
	var nf []byte
	var err error
	if p := st.NotFound; p != nil && p.HTML != nil {
		nf = p.HTML // the site's own 404.md
	} else {
		nf, err = s.theme.RenderNotFound(st, theme.Options{Dev: s.dev})
	}
	if err != nil {
		s.log.Printf("error rendering 404: %v", err)
		nf = []byte("404")
	}
	css := theme.SiteFontsCSS(st.Fonts)
	s.notFound.Store(&nf)
	s.fontsCSS.Store(&css)
	s.site.Store(st)
}

// Site returns the current snapshot.
func (s *Server) Site() *site.Site { return s.site.Load() }

// Rebuild rebuilds the snapshot from disk and notifies dev clients.
func (s *Server) Rebuild() {
	s.buildMu.Lock()
	defer s.buildMu.Unlock()
	start := time.Now()
	prev := s.site.Load()
	st, err := s.build(prev)
	if err != nil {
		s.log.Printf("error: %v", err)
		return
	}
	s.swap(st)
	errs, warns := diag.Count(st.Diags)
	s.log.Printf("rebuilt in %s · %d pages · %d error(s) · %d warning(s)", time.Since(start).Round(time.Millisecond), len(st.Pages), errs, warns)
	for _, d := range st.Diags {
		s.log.Print(d.String())
	}
	if !s.dev {
		for _, p := range st.Pages {
			if p.Stale {
				s.log.Printf("%s: serving the last valid version until the errors are fixed", p.Rel)
			} else if p.Unpublished {
				s.log.Printf("%s: not published (has errors and no previous valid version)", p.Rel)
			}
		}
	}
	s.broadcast()
}

// Watch rebuilds on file changes until ctx is done.
func (s *Server) Watch(ctx context.Context) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()
	if err := addTree(w, s.root); err != nil {
		return err
	}
	var timer *time.Timer
	pending := make(chan struct{}, 1)
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			if ignoredPath(s.root, ev.Name) {
				continue
			}
			if ev.Has(fsnotify.Create) {
				if fi, err := os.Lstat(ev.Name); err == nil && fi.IsDir() {
					_ = addTree(w, ev.Name)
				}
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(120*time.Millisecond, func() {
				select {
				case pending <- struct{}{}:
				default:
				}
			})
		case <-pending:
			s.Rebuild()
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			s.log.Printf("watcher: %v", err)
		}
	}
}

func addTree(w *fsnotify.Watcher, dir string) error {
	return filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if p != dir && (strings.HasPrefix(d.Name(), ".") || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		return w.Add(p)
	})
}

// ignoredPath filters editor and VCS noise so saves trigger one rebuild.
func ignoredPath(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return true
	}
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if strings.HasPrefix(part, ".") || part == "node_modules" {
			return true
		}
	}
	base := filepath.Base(p)
	return strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, ".tmp") || strings.HasPrefix(base, "#")
}

func (s *Server) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	s.subsMu.Lock()
	s.subs[ch] = struct{}{}
	s.subsMu.Unlock()
	return ch
}

func (s *Server) unsubscribe(ch chan struct{}) {
	s.subsMu.Lock()
	delete(s.subs, ch)
	s.subsMu.Unlock()
}

func (s *Server) broadcast() {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

const csp = "default-src 'self'; img-src 'self' https: data:; style-src 'self'; script-src 'self'; font-src 'self'; connect-src 'self'; frame-ancestors 'self'; base-uri 'self'; form-action 'self'; object-src 'none'"

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("Content-Security-Policy", csp)
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		h.Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p := r.URL.Path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	clean := path.Clean(p)
	st := s.site.Load()

	if strings.HasPrefix(clean, theme.AssetPrefix) {
		s.serveRuntime(w, r, strings.TrimPrefix(clean, theme.AssetPrefix))
		return
	}
	if clean == "/llms.txt" {
		serveBytes(w, r, "text/plain; charset=utf-8", "no-cache", st.LLMSTxt())
		return
	}
	if clean != p && clean != "/" {
		// Trailing slashes and dot segments: redirect to the canonical route.
		if _, ok := st.ByRoute[clean]; ok {
			http.Redirect(w, r, clean, http.StatusMovedPermanently)
			return
		}
	}
	if pg, ok := st.ByRaw[clean]; ok && !pg.Unpublished {
		serveBytes(w, r, "text/markdown; charset=utf-8", "no-cache", pg.Source)
		return
	}
	if pg, ok := st.ByRoute[clean]; ok && pg.HTML != nil {
		if match := r.Header.Get("If-None-Match"); match != "" && match == pg.ETag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		h.Set("ETag", pg.ETag)
		serveBytes(w, r, "text/html; charset=utf-8", "no-cache", pg.HTML)
		return
	}
	if file, ok := st.Assets[clean]; ok {
		s.serveFile(w, r, file)
		return
	}
	s.serveNotFound(w, r)
}

func (s *Server) serveNotFound(w http.ResponseWriter, r *http.Request) {
	nf := s.notFound.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
	if r.Method != http.MethodHead && nf != nil {
		_, _ = w.Write(*nf)
	}
}

func serveBytes(w http.ResponseWriter, r *http.Request, ctype, cache string, data []byte) {
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", cache)
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, file string) {
	f, err := os.Open(file)
	if err != nil {
		s.serveNotFound(w, r)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil || fi.IsDir() {
		s.serveNotFound(w, r)
		return
	}
	if strings.EqualFold(filepath.Ext(file), ".svg") {
		// SVGs from content may contain scripts; never let them run.
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}

func (s *Server) serveRuntime(w http.ResponseWriter, r *http.Request, rest string) {
	switch rest {
	case "events":
		if s.dev {
			s.serveEvents(w, r)
			return
		}
	case "site-fonts.css":
		if css := s.fontsCSS.Load(); css != nil {
			serveBytes(w, r, "text/css; charset=utf-8", "no-cache", *css)
			return
		}
	}
	if a, ok := s.theme.Lookup(rest); ok {
		cache := "public, max-age=31536000, immutable"
		if s.dev {
			cache = "no-cache"
		}
		serveBytes(w, r, a.ContentType, cache, a.Data)
		return
	}
	s.serveNotFound(w, r)
}

func (s *Server) serveEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	ch := s.subscribe()
	defer s.unsubscribe(ch)
	// The boot id lets a page that outlived a server restart (for example
	// after rebuilding the binary) notice and reload itself.
	_, _ = io.WriteString(w, "retry: 500\nevent: hello\ndata: "+s.bootID+"\n\n")
	flusher.Flush()
	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.done:
			return
		case <-ch:
			if _, err := io.WriteString(w, "event: reload\ndata: {}\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-ping.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// ListenAndServe serves until ctx is cancelled, then shuts down gracefully.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		close(s.done)
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdown)
	}
}
