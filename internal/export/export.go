// Package export writes a site snapshot as static files, for hosts without a
// Wavemark process (GitHub Pages, Netlify, Cloudflare Pages, any web server).
//
// Layout of the output folder:
//
//	index.html, guide/index.html     folder pages (index.md)
//	guide/install.html               other pages; most hosts serve /guide/install from it
//	guide/install.md, guide.md        the Markdown source of each page, as in `wavemark serve`
//	404.html                          the not-found page (404.md, or the theme's)
//	llms.txt                          the site map for language models
//	_wavemark/<hash>/...              theme assets, safe to cache forever
//	img/a.png, ...                    every other file of the content folder
//	.nojekyll                         tells GitHub Pages to publish _wavemark/
package export

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/site"
	"github.com/YutaViriato/Wavemark/internal/theme"
)

// marker identifies folders created by Wavemark, so a rebuild may clear them.
const marker = ".wavemark-build"

// Options configures an export.
type Options struct {
	Out string
	// Base is the URL path the site is published under, such as "/Wavemark"
	// for a GitHub project page. Empty means the domain root.
	Base string
}

// Result summarizes an export.
type Result struct {
	Pages, Files int
	Out          string
}

// BaseFromURL extracts the path of a baseURL ("https://x.github.io/Repo" →
// "/Repo").
func BaseFromURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return NormalizeBase(u.Path)
}

// NormalizeBase turns "Repo", "/Repo/" and "/" into "/Repo", "/Repo" and "".
func NormalizeBase(b string) string {
	b = strings.Trim(strings.TrimSpace(b), "/")
	if b == "" {
		return ""
	}
	return "/" + b
}

// Write exports the snapshot s to opts.Out.
func Write(s *site.Site, th *theme.Theme, opts Options) (Result, error) {
	out, err := filepath.Abs(opts.Out)
	if err != nil {
		return Result{}, err
	}
	if err := prepare(out, s.Root); err != nil {
		return Result{}, err
	}
	base := NormalizeBase(opts.Base)
	w := &writer{out: out, base: base}

	for _, p := range s.Pages {
		if p.HTML == nil {
			continue
		}
		w.write(htmlPath(p), w.rebase(p.HTML))
		w.write(sourcePath(p), p.Source)
		w.pages++
	}

	notFound := []byte(nil)
	if p := s.NotFound; p != nil && p.HTML != nil {
		notFound = p.HTML
	} else if nf, err := th.RenderNotFound(s, theme.Options{}); err == nil {
		notFound = nf
	}
	w.write("404.html", w.rebase(notFound))

	for urlPath, a := range th.Assets() {
		w.write(path.Join("_wavemark", urlPath), a.Data)
	}
	if f := s.Fonts; f.Sans != "" || f.Serif != "" || f.Mono != "" {
		w.write("_wavemark/site-fonts.css", w.rebaseCSS(theme.SiteFontsCSS(f)))
	}
	for urlPath, file := range s.Assets {
		w.copy(strings.TrimPrefix(urlPath, "/"), file)
	}
	w.write("llms.txt", w.rebaseText(s.LLMSTxt()))
	w.write(".nojekyll", nil)
	w.write(marker, []byte("Created by `wavemark build`. The next build may delete this folder.\n"))
	return Result{Pages: w.pages, Files: w.files, Out: out}, w.err
}

// prepare empties a previous build folder, refusing to touch anything else.
func prepare(out, root string) error {
	if rel, err := filepath.Rel(root, out); err == nil && !strings.HasPrefix(rel, "..") {
		return fmt.Errorf("the output folder %s is inside the content folder; choose a folder outside it", out)
	}
	entries, err := os.ReadDir(out)
	if errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(out, 0o755)
	}
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if _, err := os.Stat(filepath.Join(out, marker)); err != nil {
		return fmt.Errorf("%s is not empty and was not created by wavemark build; remove it or choose another --out", out)
	}
	return os.RemoveAll(out)
}

// htmlPath maps a route to its file: folder pages become index.html so that
// both /guide and /guide/ work; other pages become name.html.
func htmlPath(p *site.Page) string {
	if p.Route == "/" {
		return "index.html"
	}
	r := strings.TrimPrefix(p.Route, "/")
	if strings.EqualFold(path.Base(p.Rel), "index.md") || strings.EqualFold(path.Base(p.Rel), "readme.md") {
		return r + "/index.html"
	}
	return r + ".html"
}

func sourcePath(p *site.Page) string {
	if p.Route == "/" {
		return "index.md"
	}
	return strings.TrimPrefix(p.Route, "/") + ".md"
}

type writer struct {
	out, base    string
	pages, files int
	err          error
}

func (w *writer) target(rel string) (string, bool) {
	p := filepath.Join(w.out, filepath.FromSlash(rel))
	if r, err := filepath.Rel(w.out, p); err != nil || strings.HasPrefix(r, "..") {
		w.fail(fmt.Errorf("refusing to write outside the output folder: %s", rel))
		return "", false
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		w.fail(err)
		return "", false
	}
	return p, true
}

func (w *writer) fail(err error) {
	if w.err == nil {
		w.err = err
	}
}

func (w *writer) write(rel string, data []byte) {
	if w.err != nil {
		return
	}
	p, ok := w.target(rel)
	if !ok {
		return
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		w.fail(err)
		return
	}
	w.files++
}

func (w *writer) copy(rel, src string) {
	if w.err != nil {
		return
	}
	p, ok := w.target(rel)
	if !ok {
		return
	}
	in, err := os.Open(src)
	if err != nil {
		w.fail(err)
		return
	}
	defer in.Close()
	o, err := os.Create(p)
	if err != nil {
		w.fail(err)
		return
	}
	if _, err := io.Copy(o, in); err != nil {
		w.fail(err)
	}
	if err := o.Close(); err != nil {
		w.fail(err)
	}
	w.files++
}

// Internal URLs in generated HTML are root-absolute ("/guide", "/_wavemark/…").
// Under a base path they get the prefix. Escaped text in code blocks
// (href=&#34;/x&#34;) never matches, because real attributes are quoted.
var attrURL = regexp.MustCompile(`(\s(?:href|src))="/([^/"][^"]*)?"`)

func (w *writer) rebase(html []byte) []byte {
	if w.base == "" {
		return html
	}
	return attrURL.ReplaceAll(html, []byte(`$1="`+w.base+`/$2"`))
}

var cssURL = regexp.MustCompile(`url\("/([^/"][^"]*)"\)`)

func (w *writer) rebaseCSS(css []byte) []byte {
	if w.base == "" {
		return css
	}
	return cssURL.ReplaceAll(css, []byte(`url("`+w.base+`/$1")`))
}

// rebaseText prefixes root-relative Markdown links in llms.txt.
var mdLink = regexp.MustCompile(`\]\(/([^)]*)\)`)

func (w *writer) rebaseText(b []byte) []byte {
	if w.base == "" {
		return b
	}
	return mdLink.ReplaceAll(b, []byte(`](`+w.base+`/$1)`))
}
