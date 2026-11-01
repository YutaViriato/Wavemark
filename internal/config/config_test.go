package config

import (
	"strings"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/diag"
)

func TestSiteUnknownKeysAndSynonyms(t *testing.T) {
	d := &diag.List{File: SiteFile}
	s := LoadSite([]byte("title: X\nurl: https://x.dev\nnav:\n  - title: Guia\n    href: /guia\n  - label: API\n    link: /api\n"), d)
	if s.Title != "X" {
		t.Fatalf("title: %q", s.Title)
	}
	var hints []string
	for _, x := range d.Items {
		hints = append(hints, x.Code+": "+x.Message+" | "+x.Hint)
	}
	joined := strings.Join(hints, "\n")
	for _, want := range []string{"'url'", "baseURL", "'title'", "Navigation items use 'label'", "'link'", "Use 'href'"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
	// A misspelled key must not also produce "label is required".
	if strings.Contains(joined, "required") {
		t.Errorf("duplicate error:\n%s", joined)
	}
}

func TestShortNamesDoNotGetWildSuggestions(t *testing.T) {
	d := &diag.List{}
	LoadSite([]byte("foo: 1\n"), d)
	if len(d.Items) != 1 || strings.Contains(d.Items[0].Hint, "Did you mean") {
		t.Errorf("meaningless suggestion: %v", d.Items)
	}
}

func TestLogoShorthandAndMap(t *testing.T) {
	d := &diag.List{}
	s := LoadSite([]byte("logo: ./img/logo.svg\n"), d)
	if s.Logo.Src != "./img/logo.svg" || len(d.Items) != 0 {
		t.Errorf("logo shorthand: %+v %v", s.Logo, d.Items)
	}
	d = &diag.List{}
	LoadSite([]byte("logo:\n  text: Acme\n  alt: x\n"), d)
	if len(d.Items) != 1 || d.Items[0].Code != diag.CodeConfigUnknownKey || d.Items[0].Line != 3 {
		t.Errorf("unknown key in logo: %v", d.Items)
	}
}

func TestTypeErrorsAreReadable(t *testing.T) {
	d := &diag.List{}
	LoadFolder([]byte("order: desc\n"), d)
	if len(d.Items) != 1 || strings.Contains(d.Items[0].Message, "[]string") || !strings.Contains(d.Items[0].Message, "list") {
		t.Errorf("message: %v", d.Items)
	}
}

func TestFrontmatterDatesAndCascade(t *testing.T) {
	d := &diag.List{}
	fm := LoadFrontmatter([]byte("date: 2026-03-14\nauthor: Ada\nlayout: article\n"), 1, d)
	if len(d.Items) != 0 {
		t.Fatalf("diagnostics: %v", d.Items)
	}
	toc := false
	s := Resolve(Defaults{TOC: &toc}, fm)
	if s.Date.Format("2006-01-02") != "2026-03-14" || s.Author != "Ada" || s.Layout != "article" || s.TOC {
		t.Errorf("settings: %+v", s)
	}
	d = &diag.List{}
	LoadFrontmatter([]byte("title: x\ndate: 14/03/2026\n"), 1, d)
	if len(d.Items) != 1 || d.Items[0].Code != diag.CodeDate || d.Items[0].Line != 3 {
		t.Errorf("invalid date: %v", d.Items)
	}
}

func TestSplitFrontmatter(t *testing.T) {
	d := &diag.List{}
	fm, body, yo, bo := SplitFrontmatter([]byte("---\ntitle: a\n---\n# Body\n"), d)
	if string(fm) != "title: a\n" || string(body) != "# Body\n" || yo != 1 || bo != 3 {
		t.Errorf("split: %q %q %d %d", fm, body, yo, bo)
	}
	d = &diag.List{}
	SplitFrontmatter([]byte("---\ntitle: a\n"), d)
	if len(d.Items) != 1 {
		t.Errorf("unclosed frontmatter: %v", d.Items)
	}
}
