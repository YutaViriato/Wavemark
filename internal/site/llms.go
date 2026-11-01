package site

import (
	"strings"
)

// LLMSTxt renders /llms.txt (https://llmstxt.org): the site map for language
// models, pointing at the Markdown source of every visible page.
func (s *Site) LLMSTxt() []byte {
	var b strings.Builder
	title := s.Config.Title
	if title == "" {
		title = "Wavemark"
	}
	b.WriteString("# " + title + "\n\n")
	if s.Config.Description != "" {
		b.WriteString("> " + oneLine(s.Config.Description) + "\n\n")
	}
	url := func(p *Page) string {
		raw := p.Route + ".md"
		if p.Route == "/" {
			raw = "/index.md"
		}
		return s.Config.BaseURL + raw
	}
	item := func(p *Page) {
		if p.Unpublished || p == s.NotFound {
			return
		}
		b.WriteString("- [" + p.Title + "](" + url(p) + ")")
		if p.Summary != "" {
			b.WriteString(": " + oneLine(p.Summary))
		}
		b.WriteString("\n")
	}
	if home := s.ByRoute["/"]; home != nil {
		item(home)
	}
	var loose []*Page
	for _, c := range s.Nav.Children {
		if c.Page != nil {
			loose = append(loose, c.Page)
		}
	}
	for _, p := range loose {
		item(p)
	}
	var section func(n *NavNode, level int)
	section = func(n *NavNode, level int) {
		b.WriteString("\n" + strings.Repeat("#", min(level, 6)) + " " + n.Title + "\n\n")
		var subs []*NavNode
		for _, c := range n.Children {
			if c.IsSection() {
				subs = append(subs, c)
			} else {
				item(c.Page)
			}
		}
		for _, sub := range subs {
			section(sub, level+1)
		}
	}
	for _, c := range s.Nav.Children {
		if c.IsSection() {
			section(c, 2)
		}
	}
	return []byte(b.String())
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }
