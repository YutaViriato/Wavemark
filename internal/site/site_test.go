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
		"wavemark.yaml":      "title: Site\nnav:\n  - label: Guia\n    href: ./guia/index.md\n",
		"index.md":           "# Home\n\n[Instalar](./guia/instalacao.md#requisitos)\n",
		"guia/_meta.yaml":    "title: Guia\norder: [instalacao]\n",
		"guia/index.md":      "# Visão geral\n",
		"guia/instalacao.md": "# Instalação\n\n## Requisitos\n\n[volta](../index.md) [img](./a.png)\n",
		"guia/avancado.md":   "---\ntitle: Avançado\n---\n[ruim](./nada.md) [ancora](./instalacao.md#requisito)\n",
		"guia/a.png":         "png",
	}, Options{Dev: true})

	if s.ByRoute["/guia/instalacao"] == nil || s.ByRoute["/guia"] == nil || s.ByRoute["/"] == nil {
		t.Fatalf("rotas: %v", s.ByRoute)
	}
	if got := s.TopNav[0].Href; got != "/guia" {
		t.Errorf("nav href = %q", got)
	}
	sec := s.Nav.Children[0]
	if sec.Title != "Guia" || sec.Children[0].Page.Route != "/guia" || sec.Children[1].Page.Route != "/guia/instalacao" || sec.Children[2].Page.Route != "/guia/avancado" {
		t.Errorf("ordem da navegação: %+v", sec.Children)
	}
	home := s.ByRoute["/"]
	if !strings.Contains(home.Body, `href="/guia/instalacao#requisitos"`) {
		t.Errorf("link reescrito: %s", home.Body)
	}
	all := codes(s.Diags)
	if !strings.Contains(all, "link.broken") || !strings.Contains(all, "#requisitos?") {
		t.Errorf("diagnósticos de link:\n%s", all)
	}
	if strings.Count(all, "error") != 2 {
		t.Errorf("esperava exatamente 2 erros:\n%s", all)
	}
}

func TestPageListAndDates(t *testing.T) {
	s := build(t, map[string]string{
		"blog/_meta.yaml": "title: Blog\nsort: date\ndefaults:\n  layout: article\n",
		"blog/index.md":   "# Blog\n\n<PageList />\n",
		"blog/velho.md":   "---\ntitle: Velho\ndate: 2025-01-02\n---\nx\n",
		"blog/novo.md":    "---\ntitle: Novo\ndate: 2026-05-06\nauthor: Ana\n---\nx\n",
		"blog/meio.md":    "---\ntitle: Meio\ndate: 2025-09-01\nhidden: true\n---\nx\n",
		"x.md":            "# X\n\n<PageList dir=\"./nao-existe\" />\n",
	}, Options{Dev: true})
	body := s.ByRoute["/blog"].Body
	iNovo, iMeio, iVelho := strings.Index(body, "Novo"), strings.Index(body, "Meio"), strings.Index(body, "Velho")
	if iNovo < 0 || iMeio < 0 || iVelho < 0 || !(iNovo < iMeio && iMeio < iVelho) {
		t.Errorf("ordem por data (oculta incluída):\n%s", body)
	}
	if !strings.Contains(body, `datetime="2026-05-06"`) || !strings.Contains(body, "Ana") {
		t.Errorf("metadados:\n%s", body)
	}
	if s.ByRoute["/blog/novo"].Settings.Layout != "article" {
		t.Error("cascata de layout")
	}
	var nav []*NavNode
	for _, c := range s.Nav.Children {
		if c.Title == "Blog" {
			nav = c.Children
		}
	}
	if len(nav) != 3 || nav[0].Title != "Blog" || nav[1].Title != "Novo" || nav[2].Title != "Velho" {
		t.Fatalf("sort: date na navegação: %+v", nav)
	}
	if !strings.Contains(codes(s.Diags), "nao-existe") {
		t.Errorf("pasta inexistente não reportada:\n%s", codes(s.Diags))
	}
}

func TestRouteConflict(t *testing.T) {
	s := build(t, map[string]string{"a/index.md": "# A\n", "a/README.md": "# B\n"}, Options{Dev: true})
	if !strings.Contains(codes(s.Diags), "route.conflict") {
		t.Errorf("conflito não reportado:\n%s", codes(s.Diags))
	}
}

func TestLastGoodVersion(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"index.md": "# Home\n\nversão 1\n"})
	render := func(s *Site, p *Page) ([]byte, error) { return []byte(p.Body), nil }
	first, err := Build(Options{Root: root, Render: render})
	if err != nil || first.ByRoute["/"].HTML == nil {
		t.Fatalf("primeira versão: %v", err)
	}
	write(t, root, map[string]string{"index.md": "# Home\n\n<Crad>versão 2</Crad>\n", "nova.md": "# Nova\n\n<div>x</div>\n"})
	second, err := Build(Options{Root: root, Render: render, Prev: first})
	if err != nil {
		t.Fatal(err)
	}
	home := second.ByRoute["/"]
	if !home.Stale || !strings.Contains(string(home.HTML), "versão 1") {
		t.Errorf("devia manter a versão 1: stale=%v %s", home.Stale, home.HTML)
	}
	if !second.ByRoute["/nova"].Unpublished {
		t.Error("página nova com erro devia ficar fora do ar")
	}
}

func TestSymlinksAndDotfilesAreIgnored(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"index.md": "# Home\n", ".env": "segredo", ".git/config": "x"})
	if err := os.Symlink("/etc/hostname", filepath.Join(root, "host.txt")); err != nil {
		t.Skip("symlinks indisponíveis")
	}
	s, err := Build(Options{Root: root, Dev: true})
	if err != nil {
		t.Fatal(err)
	}
	for k := range s.Assets {
		t.Errorf("asset inesperado: %s", k)
	}
}
