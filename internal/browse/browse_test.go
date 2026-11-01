package browse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/theme"
)

func write(t *testing.T, root, rel, data string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuild(t *testing.T) {
	root := t.TempDir()
	write(t, root, "README.md", "# Project\n\nSee [the plan](docs/plan.md#next) and [docs](docs/).\n")
	write(t, root, "CLAUDE.md", "No heading here.\n")
	write(t, root, "docs/plan.md", "---\ntitle: The plan\nstatus: draft\n---\n\n## Next\n")
	write(t, root, ".kiro/specs/a/requirements.md", "# Reqs\n")
	write(t, root, "docs/img/a.png", "png")
	write(t, root, ".env", "SECRET=1")
	write(t, root, "main.go", "package main")
	write(t, root, "node_modules/x/README.md", "# dependency")
	if err := os.Symlink(filepath.Join(root, "README.md"), filepath.Join(root, "link.md")); err != nil {
		t.Fatal(err)
	}

	th, err := theme.New()
	if err != nil {
		t.Fatal(err)
	}
	s, err := Build(Options{Root: root, Theme: th, All: true})
	if err != nil {
		t.Fatal(err)
	}
	var rels []string
	for _, p := range s.Pages {
		rels = append(rels, p.Rel)
	}
	if got := strings.Join(rels, ","); got != ".kiro/specs/a/requirements.md,CLAUDE.md,README.md,docs/plan.md" {
		t.Fatalf("pages = %s", got)
	}
	for _, route := range []string{"/", "/docs", "/.kiro", "/.kiro/specs/a", "/README.md", "/docs/plan.md"} {
		if p := s.ByRoute[route]; p == nil || p.HTML == nil {
			t.Errorf("route %s missing", route)
		}
	}
	if _, ok := s.ByRaw[SourcePrefix+"docs/plan.md"]; !ok {
		t.Error("raw source missing")
	}
	if _, ok := s.Assets["/docs/img/a.png"]; !ok {
		t.Error("image not served")
	}
	for _, secret := range []string{"/.env", "/main.go"} {
		if _, ok := s.Assets[secret]; ok {
			t.Errorf("%s must not be served", secret)
		}
	}
	readme := string(s.ByRoute["/README.md"].HTML)
	if !strings.Contains(readme, `href="/docs/plan.md#next"`) || !strings.Contains(readme, `href="/docs"`) {
		t.Errorf("links not resolved")
	}
	if p := s.ByRoute["/CLAUDE.md"]; p.Title != "CLAUDE.md" || !strings.Contains(string(p.HTML), "Instructions") {
		t.Errorf("CLAUDE.md: title %q, badge missing", p.Title)
	}
	plan := s.ByRoute["/docs/plan.md"]
	if plan.Title != "The plan" || !strings.Contains(string(plan.HTML), "<dt>status</dt><dd>draft</dd>") {
		t.Errorf("frontmatter not shown: %q", plan.Title)
	}
	if s.Version == "" {
		t.Error("no version")
	}
}

func TestRelevant(t *testing.T) {
	for rel, want := range map[string]bool{"a.md": true, "docs/x.MD": true, "img.png": true, "newdir": true, "main.go": false, "a.md.swp": false, ".gitignore": true} {
		if got := Relevant(rel); got != want {
			t.Errorf("Relevant(%q) = %v", rel, got)
		}
	}
}
