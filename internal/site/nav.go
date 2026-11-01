package site

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/YutaViriato/Wavemark/internal/config"
	"github.com/YutaViriato/Wavemark/internal/diag"
)

// NavNode is an entry of the sidebar: a page or a section (folder).
type NavNode struct {
	Title    string
	Page     *Page // nil for sections
	Dir      string
	Children []*NavNode
}

// IsSection reports whether the node is a folder.
func (n *NavNode) IsSection() bool { return n.Page == nil }

// Contains reports whether p is n or one of its descendants.
func (n *NavNode) Contains(p *Page) bool {
	if n.Page == p {
		return true
	}
	for _, c := range n.Children {
		if c.Contains(p) {
			return true
		}
	}
	return false
}

type navItem struct {
	node  *NavNode
	slug  string
	order int
	has   bool
}

func (s *Site) buildNav(pages []*Page, folders map[string]config.Folder) *NavNode {
	byDir := map[string][]*Page{}
	subdirs := map[string][]string{}
	for _, p := range pages {
		if s.ByRoute[p.Route] != p {
			continue // route conflict; reported elsewhere
		}
		byDir[p.dir] = append(byDir[p.dir], p)
	}
	for dir := range folders {
		if dir != "." {
			subdirs[path.Dir(dir)] = append(subdirs[path.Dir(dir)], dir)
		}
	}

	var build func(dir string, crumbs []Crumb) *NavNode
	build = func(dir string, crumbs []Crumb) *NavNode {
		f := folders[dir]
		node := &NavNode{Dir: dir, Title: f.Title}
		var index *Page
		var items []navItem
		for _, p := range byDir[dir] {
			if p.slug == "index" {
				index = p
				continue
			}
			if p.Settings.Hidden {
				continue
			}
			items = append(items, navItem{node: &NavNode{Title: p.Title, Page: p}, slug: p.slug, order: p.Settings.Order, has: p.Settings.HasOrder})
		}
		if node.Title == "" {
			switch {
			case dir == ".":
				node.Title = s.Config.Title
			case index != nil && index.Settings.Title != "":
				node.Title = index.Title
			default:
				node.Title = prettify(path.Base(dir))
			}
		}
		here := crumbs
		if dir != "." {
			c := Crumb{Title: node.Title}
			if index != nil {
				c.Route = index.Route
			}
			here = append(append([]Crumb{}, crumbs...), c)
		}
		for _, p := range byDir[dir] {
			p.Crumbs = here
			if p == index {
				p.Crumbs = crumbs // a section's index page is the section itself
			}
		}
		for _, sub := range subdirs[dir] {
			child := build(sub, here)
			if len(child.Children) == 0 {
				continue
			}
			items = append(items, navItem{node: child, slug: path.Base(sub)})
		}

		// Explicit order from _meta.yaml first, then frontmatter order, then title.
		pos := map[string]int{}
		for i, name := range f.Order {
			pos[strings.TrimSuffix(name, ".md")] = i
		}
		known := map[string]bool{}
		for _, it := range items {
			known[it.slug] = true
		}
		for _, name := range f.Order {
			n := strings.TrimSuffix(name, ".md")
			if !known[n] && !(index != nil && n == "index") {
				s.navDiags = append(s.navDiags, diag.Diagnostic{File: path.Join(dir, config.FolderFile), Severity: diag.Warning, Code: diag.CodeConfigValue, Message: fmt.Sprintf("order: %q não corresponde a nenhuma página ou subpasta visível", name), Hint: "Use o nome do arquivo sem .md ou o nome da subpasta."})
			}
		}
		sort.SliceStable(items, func(i, j int) bool {
			a, b := items[i], items[j]
			pa, oka := pos[a.slug]
			pb, okb := pos[b.slug]
			if oka != okb {
				return oka
			}
			if oka {
				return pa < pb
			}
			if f.Sort == "date" {
				da, db := navDate(a.node), navDate(b.node)
				if !da.Equal(db) {
					return da.After(db) // newest first; undated last
				}
			}
			if a.has != b.has {
				return a.has
			}
			if a.has && a.order != b.order {
				return a.order < b.order
			}
			// Pages before sections, then alphabetical.
			if a.node.IsSection() != b.node.IsSection() {
				return !a.node.IsSection()
			}
			return strings.ToLower(a.node.Title) < strings.ToLower(b.node.Title)
		})
		if index != nil && dir != "." && !index.Settings.Hidden {
			node.Children = append(node.Children, &NavNode{Title: index.Title, Page: index})
		}
		for _, it := range items {
			node.Children = append(node.Children, it.node)
		}
		return node
	}
	return build(".", nil)
}

// flatten lists pages in navigation order followed by pages outside the
// navigation (hidden ones and the home page).
func (s *Site) flatten(pages []*Page) []*Page {
	var out []*Page
	seen := map[*Page]bool{}
	if home := s.ByRoute["/"]; home != nil {
		out = append(out, home)
		seen[home] = true
	}
	var walk func(n *NavNode)
	walk = func(n *NavNode) {
		if n.Page != nil && !seen[n.Page] {
			out = append(out, n.Page)
			seen[n.Page] = true
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(s.Nav)
	s.navOrder = out
	for _, p := range pages {
		if s.ByRoute[p.Route] == p && !seen[p] {
			out = append(out, p)
			seen[p] = true
		}
	}
	return out
}

func (s *Site) linkPrevNext() {
	// Neighbours share a layout: docs pages link to docs pages, articles to articles.
	byLayout := map[string][]*Page{}
	for _, p := range s.navOrder {
		if !p.Settings.Hidden && (p.Settings.Layout == "docs" || p.Settings.Layout == "article") {
			byLayout[p.Settings.Layout] = append(byLayout[p.Settings.Layout], p)
		}
	}
	for _, visible := range byLayout {
		for i, p := range visible {
			if i > 0 {
				p.Prev = visible[i-1]
			}
			if i+1 < len(visible) {
				p.Next = visible[i+1]
			}
		}
	}
}

func navDate(n *NavNode) time.Time {
	if n.Page != nil {
		return n.Page.Settings.Date
	}
	return time.Time{}
}
