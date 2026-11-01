package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/diag"
)

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func build(t *testing.T, files map[string]string, opts Options) *Site {
	t.Helper()
	root := t.TempDir()
	write(t, root, files)
	opts.Root = root
	s, err := Build(opts)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func codes(ds []diag.Diagnostic) string {
	var b strings.Builder
	for _, d := range ds {
		b.WriteString(d.String() + "\n")
	}
	return b.String()
}

func TestRoutesNavAndLinks(t *testing.T) {
	s := build(t, map[string]string{
		"wavemark.yaml":     "title: Site\nnav:\n  - label: Guide\n    href: ./guide/index.md\n",
		"index.md":          "# Home\n\n[Install](./guide/install.md#requirements)\n",
		"guide/_meta.yaml":  "title: Guide\norder: [install]\n",
		"guide/index.md":    "# Overview\n",
		"guide/install.md":  "# Install\n\n## Requirements\n\n[back](../index.md) [img](./a.png)\n",
		"guide/advanced.md": "---\ntitle: Advanced\n---\n[bad](./nothing.md) [anchor](./install.md#requirement)\n",
		"guide/a.png":       "png",
	}, Options{Dev: true})

	if s.ByRoute["/guide/install"] == nil || s.ByRoute["/guide"] == nil || s.ByRoute["/"] == nil {
		t.Fatalf("routes: %v", s.ByRoute)
	}
	if got := s.TopNav[0].Href; got != "/guide" {
		t.Errorf("nav href = %q", got)
	}
	sec := s.Nav.Children[0]
	if sec.Title != "Guide" || sec.Children[0].Page.Route != "/guide" || sec.Children[1].Page.Route != "/guide/install" || sec.Children[2].Page.Route != "/guide/advanced" {
		t.Errorf("navigation order: %+v", sec.Children)
	}
	home := s.ByRoute["/"]
	if !strings.Contains(home.Body, `href="/guide/install#requirements"`) {
		t.Errorf("rewritten link: %s", home.Body)
	}
	all := codes(s.Diags)
	if !strings.Contains(all, "link.broken") || !strings.Contains(all, "#requirements?") {
		t.Errorf("link diagnostics:\n%s", all)
	}
	if strings.Count(all, "error") != 2 {
		t.Errorf("expected exactly 2 errors:\n%s", all)
	}
}

func TestPageListAndDates(t *testing.T) {
	s := build(t, map[string]string{
		"blog/_meta.yaml": "title: Blog\nsort: date\ndefaults:\n  layout: article\n",
		"blog/index.md":   "# Blog\n\n<PageList />\n",
		"blog/old.md":     "---\ntitle: Old\ndate: 2025-01-02\n---\nx\n",
		"blog/new.md":     "---\ntitle: New\ndate: 2026-05-06\nauthor: Ada\n---\nx\n",
		"blog/middle.md":  "---\ntitle: Middle\ndate: 2025-09-01\nhidden: true\n---\nx\n",
		"x.md":            "# X\n\n<PageList dir=\"./missing\" />\n",
	}, Options{Dev: true})
	body := s.ByRoute["/blog"].Body
	iNew, iMiddle, iOld := strings.Index(body, ">New<"), strings.Index(body, ">Middle<"), strings.Index(body, ">Old<")
	if iNew < 0 || iMiddle < 0 || iOld < 0 || !(iNew < iMiddle && iMiddle < iOld) {
		t.Errorf("date order (hidden page included):\n%s", body)
	}
	if !strings.Contains(body, `datetime="2026-05-06"`) || !strings.Contains(body, "Ada") {
		t.Errorf("metadata:\n%s", body)
	}
	if s.ByRoute["/blog/new"].Settings.Layout != "article" {
		t.Error("layout cascade")
	}
	var nav []*NavNode
	for _, c := range s.Nav.Children {
		if c.Title == "Blog" {
			nav = c.Children
		}
	}
	if len(nav) != 3 || nav[0].Title != "Blog" || nav[1].Title != "New" || nav[2].Title != "Old" {
		t.Fatalf("sort: date in the navigation: %+v", nav)
	}
	if !strings.Contains(codes(s.Diags), "missing") {
		t.Errorf("missing folder not reported:\n%s", codes(s.Diags))
	}
}

func TestRouteConflict(t *testing.T) {
	s := build(t, map[string]string{"a/index.md": "# A\n", "a/README.md": "# B\n"}, Options{Dev: true})
	if !strings.Contains(codes(s.Diags), "route.conflict") {
		t.Errorf("conflict not reported:\n%s", codes(s.Diags))
	}
}

func TestLastGoodVersion(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"index.md": "# Home\n\nversion 1\n"})
	render := func(s *Site, p *Page) ([]byte, error) { return []byte(p.Body), nil }
	first, err := Build(Options{Root: root, Render: render})
	if err != nil || first.ByRoute["/"].HTML == nil {
		t.Fatalf("first version: %v", err)
	}
	write(t, root, map[string]string{"index.md": "# Home\n\n<Crad>version 2</Crad>\n", "new.md": "# New\n\n<div>x</div>\n"})
	second, err := Build(Options{Root: root, Render: render, Prev: first})
	if err != nil {
		t.Fatal(err)
	}
	home := second.ByRoute["/"]
	if !home.Stale || !strings.Contains(string(home.HTML), "version 1") {
		t.Errorf("should keep version 1: stale=%v %s", home.Stale, home.HTML)
	}
	if !second.ByRoute["/new"].Unpublished {
		t.Error("a new page with errors should not be published")
	}
}

func TestSymlinksAndDotfilesAreIgnored(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"index.md": "# Home\n", ".env": "secret", ".git/config": "x"})
	if err := os.Symlink("/etc/hostname", filepath.Join(root, "host.txt")); err != nil {
		t.Skip("symlinks unavailable")
	}
	s, err := Build(Options{Root: root, Dev: true})
	if err != nil {
		t.Fatal(err)
	}
	for k := range s.Assets {
		t.Errorf("unexpected asset: %s", k)
	}
}
