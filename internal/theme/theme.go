// Package theme renders pages with Wavemark's built-in theme and serves its
// embedded assets (CSS, JS, fonts) under content-hashed URLs.
package theme

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"mime"
	"path"
	"strings"
	"sync"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/icons"
	"github.com/YutaViriato/Wavemark/internal/markdown"
	"github.com/YutaViriato/Wavemark/internal/site"
)

//go:embed assets templates
var files embed.FS

// AssetPrefix is the URL prefix of theme assets and runtime endpoints.
const AssetPrefix = "/_wavemark/"

// Asset is an embedded file.
type Asset struct {
	Name        string
	Data        []byte
	ContentType string
	Hash        string
}

// Theme holds parsed templates and hashed assets.
type Theme struct {
	tmpl   *template.Template
	byLang sync.Map // lang → *template.Template with the "t" func bound

	navMu   sync.Mutex
	navSite *site.Site
	navHTML string
	assets  map[string]*Asset // by name, e.g. "wavemark.css"
	byURL   map[string]*Asset // by URL path after AssetPrefix, e.g. "3f2a…/wavemark.css"
}

// New loads the embedded theme.
func New() (*Theme, error) {
	t := &Theme{assets: map[string]*Asset{}, byURL: map[string]*Asset{}}
	err := fs.WalkDir(files, "assets", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := files.ReadFile(p)
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(p, "assets/")
		if strings.HasSuffix(name, ".txt") {
			return nil // licenses ship in the source tree, not over HTTP
		}
		sum := sha256.Sum256(data)
		a := &Asset{Name: name, Data: data, Hash: hex.EncodeToString(sum[:6]), ContentType: contentType(name)}
		t.assets[name] = a
		t.byURL[a.Hash+"/"+name] = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	// CSS references fonts by name; rewrite to hashed URLs. They are relative
	// (../<hash>/fonts/x.woff2) so the CSS works under any base path.
	if css := t.assets["wavemark.css"]; css != nil {
		out := string(css.Data)
		for name, a := range t.assets {
			if strings.HasPrefix(name, "fonts/") {
				out = strings.ReplaceAll(out, "url("+name+")", "url(../"+a.Hash+"/"+name+")")
			}
		}
		delete(t.byURL, css.Hash+"/"+css.Name)
		css.Data = []byte(out)
		sum := sha256.Sum256(css.Data)
		css.Hash = hex.EncodeToString(sum[:6])
		t.byURL[css.Hash+"/"+css.Name] = css
	}

	funcs := template.FuncMap{
		"asset": t.URL,
		"icon":  icons.HTML,
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"t":     func(string) string { return "" }, // replaced per render
		"activeTop": func(p *site.Page, href string) bool {
			if p == nil || href == "" || !strings.HasPrefix(href, "/") {
				return false
			}
			if href == "/" {
				return p.Route == "/"
			}
			return p.Route == href || strings.HasPrefix(p.Route, strings.TrimSuffix(href, "/")+"/")
		},
	}
	tmpl, err := template.New("").Funcs(funcs).ParseFS(files, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}
	t.tmpl = tmpl
	return t, nil
}

func contentType(name string) string {
	switch path.Ext(name) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".woff2":
		return "font/woff2"
	}
	if ct := mime.TypeByExtension(path.Ext(name)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// URL returns the hashed URL of an embedded asset.
func (t *Theme) URL(name string) string {
	a, ok := t.assets[name]
	if !ok {
		return AssetPrefix + "missing/" + name
	}
	return AssetPrefix + a.Hash + "/" + name
}

// Assets returns every embedded asset keyed by its URL path after
// AssetPrefix, for static export.
func (t *Theme) Assets() map[string]*Asset { return t.byURL }

// Lookup finds an asset by the URL path after AssetPrefix.
func (t *Theme) Lookup(p string) (*Asset, bool) {
	a, ok := t.byURL[p]
	return a, ok
}

type pageData struct {
	Site        *site.Site
	Page        *site.Page
	Body        template.HTML
	Layout      string
	DocTitle    string
	Description string
	LogoText    string
	RawURL      string
	Sidebar     bool
	ShowTOC     bool
	Dev         bool
	Diags       []diag.Diagnostic
	CustomFonts string
	Nav         template.HTML
	Date        *dateView
	Chapters    []markdown.Heading
}

type dateView struct{ ISO, Text string }

// Options for rendering.
type Options struct {
	Dev bool
}

func (t *Theme) data(s *site.Site, p *site.Page, opts Options) pageData {
	d := pageData{Site: s, Page: p, Dev: opts.Dev, Layout: "docs"}
	d.LogoText = s.Config.Logo.Text
	if d.LogoText == "" {
		d.LogoText = s.Config.Title
	}
	if d.LogoText == "" {
		d.LogoText = "Wavemark"
	}
	if f := s.Fonts; f.Sans != "" || f.Serif != "" || f.Mono != "" {
		sum := sha256.Sum256([]byte(f.Sans + f.Serif + f.Mono))
		d.CustomFonts = hex.EncodeToString(sum[:4])
	}
	if p == nil {
		d.DocTitle = s.Config.Title
		return d
	}
	d.Layout = p.Settings.Layout
	d.Body = template.HTML(p.Body) // #nosec: produced by the Markdown renderer, which escapes text and drops raw HTML
	d.Description = p.Summary
	if d.Description == "" {
		d.Description = s.Config.Description
	}
	switch {
	case s.Config.Title == "" || p.Title == s.Config.Title:
		d.DocTitle = p.Title
	default:
		d.DocTitle = p.Title + " · " + s.Config.Title
	}
	if p.Route == "/" {
		d.RawURL = "/index.md"
	} else {
		d.RawURL = p.Route + ".md"
	}
	d.Sidebar = d.Layout == "docs"
	d.ShowTOC = d.Layout == "docs" && p.Settings.TOC && len(p.TOC()) > 0
	d.Diags = p.Diags
	if !p.Settings.Date.IsZero() {
		d.Date = &dateView{ISO: p.Settings.Date.Format("2006-01-02"), Text: components.FormatDate(p.Settings.Date, s.Config.Lang)}
	}
	if d.Layout == "article" {
		for _, h := range p.Headings {
			if h.Level == 2 {
				d.Chapters = append(d.Chapters, h)
			}
		}
		d.ShowTOC = p.Settings.TOC && len(d.Chapters) > 0
	}
	if d.Sidebar {
		d.Nav = t.sidebar(s, p)
	}
	return d
}

func (t *Theme) execute(name string, s *site.Site, d pageData) ([]byte, error) {
	tmpl, err := t.forLang(s.Config.Lang, s.T)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, d); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// forLang returns the templates with the translation function bound,
// cloned once per language.
func (t *Theme) forLang(lang string, tr func(string) string) (*template.Template, error) {
	if v, ok := t.byLang.Load(lang); ok {
		return v.(*template.Template), nil
	}
	tmpl, err := t.tmpl.Clone()
	if err != nil {
		return nil, err
	}
	tmpl.Funcs(template.FuncMap{"t": tr})
	v, _ := t.byLang.LoadOrStore(lang, tmpl)
	return v.(*template.Template), nil
}

// sidebar renders the navigation tree once per snapshot; each page only
// marks its own link as current.
func (t *Theme) sidebar(s *site.Site, current *site.Page) template.HTML {
	t.navMu.Lock()
	if t.navSite != s {
		var b strings.Builder
		writeNav(&b, s.Nav)
		t.navSite, t.navHTML = s, b.String()
	}
	nav := t.navHTML
	t.navMu.Unlock()
	if current != nil {
		link := `<a class="wm-nav__link" href="` + html.EscapeString(current.Route) + `">`
		nav = strings.Replace(nav, link, link[:len(link)-1]+` aria-current="page">`, 1)
	}
	return template.HTML(nav) // #nosec: built from escaped titles and routes
}

func writeNav(b *strings.Builder, n *site.NavNode) {
	b.WriteString(`<ul class="wm-nav__list">`)
	for _, c := range n.Children {
		if c.IsSection() {
			b.WriteString(`<li class="wm-nav__section"><p class="wm-nav__heading">` + html.EscapeString(c.Title) + `</p>`)
			writeNav(b, c)
			b.WriteString("</li>")
			continue
		}
		b.WriteString(`<li><a class="wm-nav__link" href="` + html.EscapeString(c.Page.Route) + `">` + html.EscapeString(c.Title) + "</a></li>")
	}
	b.WriteString("</ul>")
}

// RenderPage renders the full HTML document of a page.
func (t *Theme) RenderPage(s *site.Site, p *site.Page, opts Options) ([]byte, error) {
	d := t.data(s, p, opts)
	name := "docs"
	switch d.Layout {
	case "landing", "article":
		name = d.Layout
	}
	return t.execute(name, s, d)
}

// RenderNotFound renders the 404 page.
func (t *Theme) RenderNotFound(s *site.Site, opts Options) ([]byte, error) {
	d := t.data(s, nil, opts)
	d.Layout = "landing"
	d.DocTitle = s.T("notFound")
	return t.execute("notfound", s, d)
}

// SiteFontsCSS returns the @font-face rules for fonts configured in
// wavemark.yaml.
func SiteFontsCSS(f site.Fonts) []byte {
	var b strings.Builder
	face := func(family, url, role string) {
		if url == "" {
			return
		}
		fmt.Fprintf(&b, "@font-face{font-family:%q;src:url(%q) format(\"woff2\");font-weight:100 900;font-display:swap}\n:root{--wm-font-%s:%q,var(--wm-font-%s-fallback)}\n", family, url, role, family, role)
	}
	face("Site Sans", f.Sans, "sans")
	face("Site Serif", f.Serif, "serif")
	face("Site Mono", f.Mono, "mono")
	return []byte(b.String())
}
