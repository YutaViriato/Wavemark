package theme

import (
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/YutaViriato/Wavemark/internal/icons"
	"github.com/YutaViriato/Wavemark/internal/site"
)

// FileView describes a Markdown file shown by 'wavemark browse'.
type FileView struct {
	Path      string // slash-separated path relative to the browsed folder
	Name      string // file name
	Route     string
	SourceURL string
	Title     string
	Summary   string
	Modified  time.Time
	Words     int
	Badges    []Badge
	Facts     []Fact // frontmatter, in source order
	TooLarge  bool
}

// Badge is a label on a file, such as "Instructions" for CLAUDE.md.
type Badge struct{ Label, Tone string }

// Fact is a frontmatter key and its value as text.
type Fact struct{ Key, Value string }

// FolderView is a folder of files in a listing.
type FolderView struct {
	Path  string // "" for the browsed folder itself
	Route string
	Files []*FileView
}

// ListingView is the directory page of 'wavemark browse'.
type ListingView struct {
	Title   string
	Path    string // folder shown, "" for the root
	Root    string // absolute path of the browsed folder
	Files   int
	Folders int
	Updated time.Time
	Recent  []*FileView
	Groups  []*FolderView
	Now     time.Time
}

type whenView struct{ Modified, Now time.Time }

type browseData struct {
	pageData
	File    *FileView
	Listing *ListingView
	Now     time.Time
}

// RenderFile renders a browsed Markdown file in the docs layout.
func (t *Theme) RenderFile(s *site.Site, p *site.Page, f *FileView, now time.Time, opts Options) ([]byte, error) {
	d := t.data(s, p, opts)
	d.RawURL = f.SourceURL
	d.Layout, d.Sidebar = "docs", true
	d.ShowTOC = len(p.TOC()) > 0
	d.Nav = t.fileSidebar(s, p)
	return t.executeAny("browse-file", s, browseData{pageData: d, File: f, Now: now})
}

// RenderListing renders a folder listing of 'wavemark browse'.
func (t *Theme) RenderListing(s *site.Site, p *site.Page, l *ListingView, opts Options) ([]byte, error) {
	d := t.data(s, p, opts)
	d.RawURL = ""
	d.Layout, d.Sidebar, d.ShowTOC = "docs", true, false
	d.Description = fmt.Sprintf("%d Markdown files in %s", l.Files, l.Root)
	d.Nav = t.fileSidebar(s, p)
	return t.executeAny("browse-listing", s, browseData{pageData: d, Listing: l, Now: l.Now})
}

// Ago formats the time between t and now the way people say it.
func Ago(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < 45*time.Second:
		return "just now"
	case d < 90*time.Second:
		return "1 minute ago"
	case d < 45*time.Minute:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()+0.5))
	case d < 90*time.Minute:
		return "1 hour ago"
	case d < 22*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()+0.5))
	case d < 36*time.Hour:
		return "yesterday"
	case d < 26*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24+0.5))
	}
	if t.Year() == now.Year() {
		return t.Format("Jan 2")
	}
	return t.Format("Jan 2, 2006")
}

// fileSidebar renders the file tree once per snapshot. Folders are
// collapsible; each page only marks its own link as current.
func (t *Theme) fileSidebar(s *site.Site, current *site.Page) template.HTML {
	t.navMu.Lock()
	if t.fileNavSite != s {
		var b strings.Builder
		writeFileNav(&b, s.Nav)
		t.fileNavSite, t.fileNavHTML = s, b.String()
	}
	nav := t.fileNavHTML
	t.navMu.Unlock()
	if current != nil {
		link := `<a class="wm-nav__link" href="` + html.EscapeString(current.Route) + `">`
		nav = strings.Replace(nav, link, link[:len(link)-1]+` aria-current="page">`, 1)
	}
	return template.HTML(nav) // #nosec: built from escaped names and routes
}

func writeFileNav(b *strings.Builder, n *site.NavNode) {
	b.WriteString(`<ul class="wm-nav__list">`)
	for _, c := range n.Children {
		if c.IsSection() {
			b.WriteString(`<li><details class="wm-nav__folder" open><summary class="wm-nav__link wm-nav__summary">` + icons.SVG("chevron-right") + `<span>` + html.EscapeString(c.Title) + `</span></summary>`)
			writeFileNav(b, c)
			b.WriteString("</details></li>")
			continue
		}
		b.WriteString(`<li><a class="wm-nav__link" href="` + html.EscapeString(c.Page.Route) + `">` + html.EscapeString(c.Title) + "</a></li>")
	}
	b.WriteString("</ul>")
}
