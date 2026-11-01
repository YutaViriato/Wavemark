package markdown

import (
	"strings"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/diag"
)

func renderPlain(t *testing.T, src string) string {
	t.Helper()
	doc := ParsePlain([]byte(src), 0)
	Analyze(doc, "docs", false, &diag.List{})
	out, err := Render(doc, RenderOptions{Plain: true})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestPlainSanitizesHTML(t *testing.T) {
	out := renderPlain(t, strings.Join([]string{
		`<script>alert(1)</script>`,
		``,
		`<img src="javascript:alert(1)" onerror="alert(2)">`,
		``,
		`<a href="javascript:alert(3)" onclick="x()">bad</a> <a href="https://example.com" style="color:red">good</a>`,
		``,
		`<div align="center" class="x" style="y"><img src="logo.png" alt="Logo" width="120" onload="z()"></div>`,
		``,
		`<iframe src="https://evil.example"></iframe>`,
		``,
		`Press <kbd>Ctrl</kbd> <span onmouseover="x()">hi</span><br/>`,
		``,
		`</div></div></main>`,
	}, "\n"))
	for _, bad := range []string{"<script", "<iframe", "javascript:", "onerror", "onclick", "onload", "onmouseover", "style=", "class=\"x\"", "</main>"} {
		if strings.Contains(out, bad) {
			t.Errorf("output contains %q:\n%s", bad, out)
		}
	}
	for _, good := range []string{
		`&lt;script&gt;`, `<a>bad</a>`, `<a href="https://example.com" rel="noopener">good</a>`,
		`<div align="center"><img src="logo.png" alt="Logo" width="120"`, `<kbd>Ctrl</kbd>`, `<span>hi</span><br>`,
		`&lt;iframe src=&#34;https://evil.example&#34;&gt;`,
	} {
		if !strings.Contains(out, good) {
			t.Errorf("output lacks %q:\n%s", good, out)
		}
	}
	if strings.Count(out, "<div") != strings.Count(out, "</div>") {
		t.Errorf("unbalanced divs:\n%s", out)
	}
}

func TestPlainUnclosedTagsAreClosed(t *testing.T) {
	out := renderPlain(t, "<details>\n<summary>More</summary>\n\nBody text.\n")
	if !strings.HasSuffix(strings.TrimSpace(out), "</details>") {
		t.Errorf("details not closed:\n%s", out)
	}
}

func TestPlainAlertsAndIndentedCode(t *testing.T) {
	out := renderPlain(t, "> [!WARNING]\n> Do not push.\n\n> [!NOTE]\n\n> A quote.\n\n    indented\n")
	if !strings.Contains(out, `wm-callout--warning`) || !strings.Contains(out, "<p>Do not push.</p>") || strings.Contains(out, "[!WARNING]") {
		t.Errorf("warning alert not converted:\n%s", out)
	}
	if !strings.Contains(out, `wm-callout--note`) {
		t.Errorf("empty note alert not converted:\n%s", out)
	}
	if !strings.Contains(out, "<blockquote>\n<p>A quote.</p>") {
		t.Errorf("plain quote changed:\n%s", out)
	}
	if !strings.Contains(out, "<code>indented\n</code>") {
		t.Errorf("indented code not rendered:\n%s", out)
	}
}

func TestPlainIgnoresComponentsAndAttributes(t *testing.T) {
	out := renderPlain(t, "## Config {json}\n\n<Card title=\"x\">\ntext\n</Card>\n")
	if !strings.Contains(out, "Config {json}") {
		t.Errorf("heading attribute syntax applied:\n%s", out)
	}
	if !strings.Contains(out, "&lt;Card title=&#34;x&#34;&gt;") || !strings.Contains(out, "text") {
		t.Errorf("component tag not shown as text:\n%s", out)
	}
}
