// Package site turns a content folder into an immutable snapshot: pages
// with their effective settings, routes, navigation, resolved links and
// rendered HTML. The server swaps snapshots atomically when files change.
package site

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/config"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/i18n"
	"github.com/YutaViriato/Wavemark/internal/markdown"
	"github.com/YutaViriato/Wavemark/internal/sketch"
)

// Site is an immutable snapshot of a content folder.
type Site struct {
	Root    string
	Config  config.Site
	Pages   []*Page          // every page, in navigation order
	ByRoute map[string]*Page // route → page
	ByRaw   map[string]*Page // "/route.md" and "/rel/path.md" → page
	Assets  map[string]string
	Nav     *NavNode
	TopNav  []NavLink
	Footer  Footer
	Fonts   Fonts
	Diags   []diag.Diagnostic
	Built   time.Time
	T       func(string) string

	byRel    map[string]*Page
	dirs     map[string]bool
	navOrder []*Page
	navDiags []diag.Diagnostic
}

// Footer is the resolved site footer.
type Footer struct {
	Text  string
	Links []NavLink
}

// Fonts are the URLs of site-provided font files ("" = theme default).
type Fonts struct{ Sans, Serif, Mono string }

// NavLink is a resolved top navigation link.
type NavLink struct {
	Label, Href string
}

// Page is one Markdown file.
type Page struct {
	Rel         string // slash-separated path relative to the root
	Abs         string
	Route       string
	Source      []byte
	Settings    config.Settings
	Title       string
	Description string // from the frontmatter; shown as the lead
	Summary     string // description or first paragraph; for meta and llms.txt
	Headings    []markdown.Heading
	IDs         map[string]bool
	Crumbs      []Crumb
	Prev, Next  *Page
	Body        string // rendered body HTML
	HTML        []byte // full document, rendered by the theme
	ETag        string
	Diags       []diag.Diagnostic
	// Stale is set when the current file has errors and the server keeps
	// serving the last valid version.
	Stale bool
	// Unpublished is set when the page has errors and no valid version.
	Unpublished bool

	dir      string
	slug     string
	doc      *markdown.Document
	analysis *markdown.Analysis
}

// Crumb is a breadcrumb entry.
type Crumb struct{ Title, Route string }

// TOC returns the headings shown in the table of contents (h2 and h3).
func (p *Page) TOC() []markdown.Heading {
	var out []markdown.Heading
	for _, h := range p.Headings {
		if h.Level <= 3 {
			out = append(out, h)
		}
	}
	return out
}

// HasErrors reports whether the page has error diagnostics.
func (p *Page) HasErrors() bool { return diag.HasErrors(p.Diags) }

// RenderFunc renders the full HTML document of a page.
type RenderFunc func(s *Site, p *Page) ([]byte, error)

// Options configures Build.
type Options struct {
	Root string
	// Dev keeps pages with errors visible (the theme shows the errors).
	// Otherwise pages with errors fall back to Prev or are unpublished.
	Dev    bool
	Prev   *Site
	Render RenderFunc
}

// Build scans the folder and produces a snapshot. The error is only for
// conditions that make a snapshot impossible (unreadable root); content
// problems are diagnostics.
func Build(opts Options) (*Site, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}
	if fi, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("pasta do site: %w", err)
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%s não é uma pasta", root)
	}
	s := &Site{Root: root, ByRoute: map[string]*Page{}, ByRaw: map[string]*Page{}, Assets: map[string]string{}, byRel: map[string]*Page{}, Built: time.Now()}

	siteDiags := &diag.List{File: config.SiteFile}
	if data, err := os.ReadFile(filepath.Join(root, config.SiteFile)); err == nil {
		s.Config = config.LoadSite(data, siteDiags)
	} else if errors.Is(err, fs.ErrNotExist) {
		s.Config = config.LoadSite(nil, siteDiags)
	} else {
		return nil, err
	}
	s.T = i18n.For(s.Config.Lang)

	sc, err := scan(root)
	if err != nil {
		return nil, err
	}
	var folderDiags []diag.Diagnostic
	s.dirs = map[string]bool{}
	for _, d := range sc.dirs {
		s.dirs[d] = true
	}
	for _, a := range sc.assets {
		s.Assets["/"+a] = filepath.Join(root, filepath.FromSlash(a))
	}

	// Folder cascade.
	folders := map[string]config.Folder{}
	cascade := map[string]config.Defaults{}
	for _, dir := range sc.dirs { // sorted, parents first
		var f config.Folder
		if sc.metas[dir] {
			rel := path.Join(dir, config.FolderFile)
			fd := &diag.List{File: rel}
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				fd.Errorf(0, 0, diag.CodeConfigInvalid, "", "não foi possível ler: %v", err)
			} else {
				f = config.LoadFolder(data, fd)
			}
			folderDiags = append(folderDiags, fd.Items...)
		}
		folders[dir] = f
		parent := s.Config.Defaults
		if dir != "." {
			parent = cascade[path.Dir(dir)]
		}
		cascade[dir] = parent.Merge(f.Defaults)
	}

	// Parse pages in parallel.
	pages := make([]*Page, len(sc.pages))
	parallelN(len(sc.pages), func(i int) {
		rel := sc.pages[i]
		pages[i] = loadPage(root, rel, cascade[path.Dir(rel)])
	})

	// Routes.
	for _, p := range pages {
		if other, ok := s.ByRoute[p.Route]; ok {
			addDiag(p, diag.Diagnostic{Line: 1, Code: diag.CodeRouteConflict, Message: fmt.Sprintf("a rota %s também é gerada por %s", p.Route, other.Rel), Hint: "Renomeie ou remova um dos arquivos (index.md e README.md na mesma pasta, ou x.md e x/index.md, geram a mesma rota)."})
			continue
		}
		s.ByRoute[p.Route] = p
		s.byRel[p.Rel] = p
		s.ByRaw["/"+p.Rel] = p
		if p.Route == "/" {
			s.ByRaw["/index.md"] = p
		} else {
			s.ByRaw[p.Route+".md"] = p
		}
	}

	// Links, fonts and top navigation need the full route index.
	for _, p := range pages {
		if s.ByRoute[p.Route] == p {
			s.checkRefs(p)
		}
	}
	s.TopNav = s.resolveLinks(s.Config.Nav, "nav", siteDiags)
	s.Footer = Footer{Text: s.Config.Footer.Text, Links: s.resolveLinks(s.Config.Footer.Links, "footer.links", siteDiags)}
	s.Fonts = s.resolveFonts(siteDiags)

	s.Nav = s.buildNav(pages, folders)
	s.Pages = s.flatten(pages)
	s.linkPrevNext()

	// Pages are independent once the route index exists: render in parallel.
	parallel(s.Pages, func(p *Page) {
		body, err := markdown.Render(p.doc, markdown.RenderOptions{URL: s.urlFor(p), T: s.T, Pages: s.pagesFor(p), Layout: p.Settings.Layout, Lang: s.Config.Lang})
		if err != nil {
			addDiag(p, diag.Diagnostic{Code: diag.CodeConfigInvalid, Message: "falha ao renderizar: " + err.Error()})
		}
		p.Body = body
		p.doc, p.analysis = nil, nil // the AST is not needed after rendering
	})

	s.applyLastGood(opts)

	if opts.Render != nil {
		var mu sync.Mutex
		var firstErr error
		parallel(s.Pages, func(p *Page) {
			if p.Stale || p.Unpublished {
				return
			}
			html, err := opts.Render(s, p)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("renderizando %s: %w", p.Rel, err)
				}
				mu.Unlock()
				return
			}
			p.HTML = html
			sum := sha256.Sum256(html)
			p.ETag = `"` + hex.EncodeToString(sum[:8]) + `"`
		})
		if firstErr != nil {
			return nil, firstErr
		}
	}

	s.Diags = append(append(siteDiags.Items, folderDiags...), s.navDiags...)
	for _, p := range pages {
		s.Diags = append(s.Diags, p.Diags...)
	}
	diag.Sort(s.Diags)
	return s, nil
}

func addDiag(p *Page, d diag.Diagnostic) {
	if d.File == "" {
		d.File = p.Rel
	}
	if d.Severity == "" {
		d.Severity = diag.Error
	}
	p.Diags = append(p.Diags, d)
}

// applyLastGood implements the serving policy for pages with errors.
func (s *Site) applyLastGood(opts Options) {
	if opts.Dev {
		return
	}
	for i, p := range s.Pages {
		if !p.HasErrors() {
			continue
		}
		var prev *Page
		if opts.Prev != nil {
			prev = opts.Prev.ByRoute[p.Route]
		}
		if prev != nil && prev.HTML != nil {
			kept := *prev
			kept.Stale = true
			kept.Diags = p.Diags
			s.Pages[i] = &kept
			s.ByRoute[p.Route] = &kept
			for k, v := range s.ByRaw {
				if v == p {
					s.ByRaw[k] = &kept
				}
			}
			continue
		}
		p.Unpublished = true
	}
}

func loadPage(root, rel string, cascade config.Defaults) *Page {
	p := &Page{Rel: rel, Abs: filepath.Join(root, filepath.FromSlash(rel)), dir: path.Dir(rel)}
	p.Route, p.slug = route(rel)
	dl := &diag.List{File: rel}
	defer func() { p.Diags = append(p.Diags, dl.Items...) }()

	src, err := os.ReadFile(p.Abs)
	if err != nil {
		dl.Errorf(0, 0, diag.CodeConfigInvalid, "", "não foi possível ler: %v", err)
		src = nil
	}
	p.Source = src
	fmData, body, yamlOffset, bodyOffset := config.SplitFrontmatter(src, dl)
	fm := config.LoadFrontmatter(fmData, yamlOffset, dl)
	p.Settings = config.Resolve(cascade, fm)

	p.doc = markdown.Parse(body, bodyOffset, dl)
	p.analysis = markdown.Analyze(p.doc, p.Settings.Layout, fm.Title != nil, dl)
	p.Headings = p.analysis.Headings
	p.IDs = p.analysis.IDs

	switch {
	case p.Settings.Title != "":
		p.Title = p.Settings.Title
	case p.analysis.Title != "":
		p.Title = p.analysis.Title
	case p.analysis.HeroTitle != "":
		p.Title = p.analysis.HeroTitle
	default:
		p.Title = prettify(p.slug)
		dl.Warnf(1, 1, diag.CodeTitleMissing, "Adicione 'title:' no frontmatter ou comece o corpo com '# Título'.", "página sem título; usando o nome do arquivo (%q)", p.Title)
	}
	if icon := p.Settings.Icon; icon != "" {
		if _, _, ok := sketch.Lookup(icon); !ok {
			hint := "Ícones disponíveis: " + strings.Join(sketch.Names(), ", ") + "."
			if sug := diag.Suggest(icon, sketch.AllNames()); sug != "" {
				hint = fmt.Sprintf("Você quis dizer %q? %s", sug, hint)
			}
			dl.Errorf(0, 0, diag.CodeIconUnknown, hint, "ícone desconhecido %q no frontmatter", icon)
			p.Settings.Icon = ""
		}
	}
	p.Description = p.Settings.Description
	p.Summary = p.Description
	if p.Summary == "" {
		p.Summary = p.analysis.Plain
	}
	return p
}

// route maps a page path to its URL route and its slug (name in its folder).
func route(rel string) (string, string) {
	dir, file := path.Split(rel)
	name := strings.TrimSuffix(file, path.Ext(file))
	dir = strings.TrimSuffix(dir, "/")
	if strings.EqualFold(name, "index") || strings.EqualFold(name, "readme") {
		if dir == "" {
			return "/", "index"
		}
		return "/" + dir, "index"
	}
	if dir == "" {
		return "/" + name, name
	}
	return "/" + dir + "/" + name, name
}

// prettify turns a file or folder name into a title: "primeiros-passos" →
// "Primeiros passos".
func prettify(name string) string {
	s := strings.NewReplacer("-", " ", "_", " ").Replace(name)
	s = strings.TrimSpace(s)
	if s == "" {
		return name
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

type scanResult struct {
	pages  []string
	assets []string
	dirs   []string
	metas  map[string]bool
}

// ignored reports names the scanner never reads or serves.
func ignored(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules"
}

func scan(root string) (*scanResult, error) {
	r := &scanResult{metas: map[string]bool{}}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if rel != "." && ignored(d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil // never follow symlinks out of the content folder
		}
		if d.IsDir() {
			r.dirs = append(r.dirs, rel)
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		switch {
		case rel == config.SiteFile:
		case d.Name() == config.FolderFile:
			r.metas[path.Dir(rel)] = true
		case strings.EqualFold(path.Ext(rel), ".md"):
			r.pages = append(r.pages, rel)
		default:
			r.assets = append(r.assets, rel)
		}
		return nil
	})
	sort.Strings(r.dirs)
	sort.Strings(r.pages)
	return r, err
}

// parallelN runs fn(0..n-1) on GOMAXPROCS workers.
func parallelN(n int, fn func(i int)) {
	workers := min(runtime.GOMAXPROCS(0), n)
	var next atomic.Int64
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= n {
					return
				}
				fn(i)
			}
		}()
	}
	wg.Wait()
}

func parallel(pages []*Page, fn func(*Page)) {
	parallelN(len(pages), func(i int) { fn(pages[i]) })
}

// pagesFor lists the pages of a folder for <PageList> on page p: the pages in
// the folder and the index pages of its direct subfolders, excluding p.
// Hidden pages are included: hidden only removes them from the sidebar.
func (s *Site) pagesFor(p *Page) func(string) []components.PageInfo {
	return func(dir string) []components.PageInfo {
		target, ok := s.resolveDir(p, dir)
		if !ok {
			return nil
		}
		pos := map[*Page]int{}
		for i, q := range s.navOrder {
			pos[q] = i
		}
		var list []*Page
		for _, q := range s.Pages {
			if q == p || q.Unpublished {
				continue
			}
			inDir := q.dir == target && q.slug != "index"
			subIndex := q.slug == "index" && q.dir != target && path.Dir(q.dir) == target
			if target == "." && q.dir != "." && q.slug == "index" && path.Dir(q.dir) == "." {
				subIndex = true
			}
			if inDir || subIndex {
				list = append(list, q)
			}
		}
		sort.SliceStable(list, func(i, j int) bool {
			pi, oki := pos[list[i]]
			pj, okj := pos[list[j]]
			if oki != okj {
				return oki
			}
			return pi < pj
		})
		out := make([]components.PageInfo, 0, len(list))
		for _, q := range list {
			out = append(out, components.PageInfo{Title: q.Title, Description: q.Summary, Route: q.Route, Date: q.Settings.Date, Author: q.Settings.Author, Icon: q.Settings.Icon})
		}
		return out
	}
}

// resolveDir maps a folder reference to a folder path relative to the root.
func (s *Site) resolveDir(from *Page, dest string) (string, bool) {
	dest = strings.TrimSpace(dest)
	var clean string
	if strings.HasPrefix(dest, "/") {
		clean = strings.TrimPrefix(path.Clean(dest), "/")
	} else {
		clean = path.Join(from.dir, dest)
	}
	if clean == "" {
		clean = "."
	}
	return clean, s.dirs[clean]
}
