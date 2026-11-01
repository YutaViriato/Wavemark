package theme

import (
	"strings"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/site"
)

// TestExampleSite keeps the documentation in site/ valid and renders every
// page through the real templates.
func TestExampleSite(t *testing.T) {
	th, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s, err := site.Build(site.Options{Root: "../../site", Render: func(s *site.Site, p *site.Page) ([]byte, error) {
		return th.RenderPage(s, p, Options{})
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range s.Diags {
		if d.Severity == diag.Error {
			t.Errorf("site/: %s", d)
		}
	}
	layouts := map[string]bool{}
	for _, p := range s.Pages {
		if p.HTML == nil {
			t.Errorf("%s não renderizou", p.Rel)
			continue
		}
		html := string(p.HTML)
		layouts[p.Settings.Layout] = true
		if !strings.Contains(html, `data-layout="`+p.Settings.Layout+`"`) || !strings.Contains(html, "</html>") {
			t.Errorf("%s: documento incompleto", p.Rel)
		}
		if strings.Contains(html, ` style="`) {
			t.Errorf("%s: atributo style bloqueado pela CSP", p.Rel)
		}
	}
	for _, l := range []string{"docs", "landing", "article"} {
		if !layouts[l] {
			t.Errorf("o site de exemplo não usa o layout %s", l)
		}
	}
	if _, err := th.RenderNotFound(s, Options{}); err != nil {
		t.Error(err)
	}
}

func TestAssetsAreHashed(t *testing.T) {
	th, err := New()
	if err != nil {
		t.Fatal(err)
	}
	css := th.URL("wavemark.css")
	if !strings.HasPrefix(css, AssetPrefix) || strings.Contains(css, "missing") {
		t.Fatalf("url: %s", css)
	}
	a, ok := th.Lookup(strings.TrimPrefix(css, AssetPrefix))
	if !ok || !strings.Contains(string(a.Data), AssetPrefix) {
		t.Error("o CSS deveria apontar para as fontes com hash")
	}
}
