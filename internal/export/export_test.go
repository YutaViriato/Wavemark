package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/site"
	"github.com/YutaViriato/Wavemark/internal/theme"
)

func TestWriteUnderBasePath(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"wavemark.yaml":    "title: T\nbaseURL: https://x.github.io/Repo\nnav:\n  - label: Guide\n    href: /guide\n",
		"index.md":         "# Home\n\n[Install](./guide/install.md) ![Logo](./img/a.png)\n\n```html\n<a href=\"/raw\">x</a>\n```\n",
		"guide/index.md":   "# Guide\n",
		"guide/install.md": "# Install\n",
		"img/a.png":        "png",
		"404.md":           "# Lost\n",
	}
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	th, err := theme.New()
	if err != nil {
		t.Fatal(err)
	}
	s, err := site.Build(site.Options{Root: root, Render: func(s *site.Site, p *site.Page) ([]byte, error) { return th.RenderPage(s, p, theme.Options{}) }})
	if err != nil {
		t.Fatal(err)
	}
	base := BaseFromURL(s.Config.BaseURL)
	if base != "/Repo" {
		t.Fatalf("base = %q", base)
	}
	out := filepath.Join(t.TempDir(), "dist")
	if _, err := Write(s, th, Options{Out: out, Base: base}); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"index.html", "guide/index.html", "guide/install.html", "guide/install.md", "guide.md", "index.md", "404.html", "llms.txt", ".nojekyll", "img/a.png"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	home, _ := os.ReadFile(filepath.Join(out, "index.html"))
	h := string(home)
	for _, want := range []string{`href="/Repo/guide/install"`, `src="/Repo/img/a.png"`, `href="/Repo/guide"`, `href="/Repo/_wavemark/`, `href="/Repo/"`} {
		if !strings.Contains(h, want) {
			t.Errorf("index.html is missing %s", want)
		}
	}
	if strings.Contains(h, `href="/guide`) {
		t.Error("an internal link was not rebased")
	}
	if !strings.Contains(h, "&lt;a href=&#34;/raw&#34;&gt;") && !strings.Contains(h, "/raw") {
		t.Error("code content changed")
	}
	nf, _ := os.ReadFile(filepath.Join(out, "404.html"))
	if !strings.Contains(string(nf), "Lost") {
		t.Error("404.html should come from 404.md")
	}
}

func TestRefusesForeignFolders(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "index.md"), []byte("# Home\n"), 0o644)
	th, _ := theme.New()
	s, err := site.Build(site.Options{Root: root, Render: func(s *site.Site, p *site.Page) ([]byte, error) { return th.RenderPage(s, p, theme.Options{}) }})
	if err != nil {
		t.Fatal(err)
	}
	foreign := t.TempDir()
	_ = os.WriteFile(filepath.Join(foreign, "keep.txt"), []byte("mine"), 0o644)
	if _, err := Write(s, th, Options{Out: foreign}); err == nil {
		t.Error("should refuse to clear a folder it did not create")
	}
	if _, err := Write(s, th, Options{Out: filepath.Join(root, "dist")}); err == nil {
		t.Error("should refuse an output folder inside the content folder")
	}
	out := filepath.Join(t.TempDir(), "dist")
	for i := 0; i < 2; i++ { // a second build replaces the first
		if _, err := Write(s, th, Options{Out: out}); err != nil {
			t.Fatalf("build %d: %v", i+1, err)
		}
	}
}

func TestNormalizeBase(t *testing.T) {
	for in, want := range map[string]string{"": "", "/": "", "Repo": "/Repo", "/Repo/": "/Repo", " /a/b ": "/a/b"} {
		if got := NormalizeBase(in); got != want {
			t.Errorf("NormalizeBase(%q) = %q, want %q", in, got, want)
		}
	}
}
