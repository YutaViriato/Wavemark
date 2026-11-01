package markdown

import (
	"slices"
	"strings"
	"testing"

	"github.com/YutaViriato/Wavemark/internal/diag"
)

type result struct {
	html  string
	codes []string
	diags []diag.Diagnostic
	a     *Analysis
}

func run(t *testing.T, src, layout string) result {
	t.Helper()
	d := &diag.List{File: "t.md"}
	doc := Parse([]byte(src), 0, d)
	a := Analyze(doc, layout, false, d)
	html, err := Render(doc, RenderOptions{Layout: layout, Lang: "en"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	var codes []string
	for _, x := range d.Items {
		codes = append(codes, x.Code)
	}
	return result{html: html, codes: codes, diags: d.Items, a: a}
}

func (r result) has(t *testing.T, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(r.html, s) {
			t.Errorf("HTML is missing %q:\n%s", s, r.html)
		}
	}
}

func (r result) lacks(t *testing.T, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if strings.Contains(r.html, s) {
			t.Errorf("HTML should not contain %q:\n%s", s, r.html)
		}
	}
}

func (r result) clean(t *testing.T) {
	t.Helper()
	for _, d := range r.diags {
		if d.Severity == diag.Error {
			t.Errorf("unexpected error: %s", d)
		}
	}
}

func (r result) code(t *testing.T, code string, line int) {
	t.Helper()
	for _, d := range r.diags {
		if d.Code == code && (line == 0 || d.Line == line) {
			return
		}
	}
	t.Errorf("expected %s on line %d; diagnostics: %v", code, line, r.diags)
}

func TestNestedComponentsAndIndentation(t *testing.T) {
	r := run(t, `<Tabs>
  <Tab title="A">
    Some **bold** text.

    - item
      - nested
  </Tab>
  <Tab title="B">One line</Tab>
</Tabs>
`, "docs")
	r.clean(t)
	r.has(t, `role="tab"`, `<strong>bold</strong>`, `<li>nested</li>`, `<p>One line</p>`)
	if strings.Count(r.html, `role="tabpanel"`) != 2 {
		t.Errorf("expected 2 panels:\n%s", r.html)
	}
}

func TestSameNameNesting(t *testing.T) {
	r := run(t, "<Columns>\n<Columns cols={1}>\n<Card title=\"in\">x</Card>\n</Columns>\n<Card title=\"out\">y</Card>\n</Columns>\n", "docs")
	r.clean(t)
	if strings.Count(r.html, `wm-columns--`) != 2 {
		t.Errorf("expected two Columns:\n%s", r.html)
	}
	// "out" must be inside the outer grid, after the inner one closed.
	if i, j := strings.Index(r.html, "wm-columns--1"), strings.Index(r.html, ">out<"); i < 0 || j < i {
		t.Errorf("wrong order:\n%s", r.html)
	}
}

func TestClosingTagInsideFenceIsCode(t *testing.T) {
	r := run(t, "<Tabs>\n<Tab title=\"x\">\n```html\n</Tab>\n```\n</Tab>\n</Tabs>\n", "docs")
	r.clean(t)
	r.has(t, `<span class="hl-nt">Tab</span>`, `role="tabpanel"`)
}

func TestSelfClosingInsideContainer(t *testing.T) {
	r := run(t, "<Hero title=\"Hi\">\nSub.\n<Button href=\"/\" label=\"Go\" />\n<Button href=\"/b\">Other</Button>\n</Hero>\n", "landing")
	r.clean(t)
	r.has(t, `<div class="wm-actions"><a class="wm-button wm-button--primary" href="/">Go</a><a class="wm-button wm-button--primary" href="/b">Other</a></div>`)
	r.lacks(t, "<p>Other</p>")
}

func TestMultilineOpeningTag(t *testing.T) {
	r := run(t, "<Card\n  title=\"Long\"\n  icon=\"rocket\"\n>\nBody.\n</Card>\n", "docs")
	r.clean(t)
	r.has(t, ">Long<", "wm-sketch", "<p>Body.</p>")
}

func TestComponentErrors(t *testing.T) {
	r := run(t, `<Note>
never closed

</Steps>

<Crad title="x">a</Crad>

<Callout type="caution">x</Callout>

<Card title={x + 1}>a</Card>

<Card titl="a">b</Card>

<Accordion title="x" open="yes">y</Accordion>
`, "docs")
	r.code(t, diag.CodeComponentUnclosed, 1)
	r.code(t, diag.CodeComponentStray, 4)
	r.code(t, diag.CodeComponentUnknown, 6)
	r.code(t, diag.CodeAttrValue, 8)
	r.code(t, diag.CodeComponentSyntax, 10)
	r.code(t, diag.CodeAttrUnknown, 12)
	r.code(t, diag.CodeAttrMissing, 12)
	r.code(t, diag.CodeAttrValue, 14)
	for _, d := range r.diags {
		if d.Code == diag.CodeComponentUnknown && !strings.Contains(d.Hint, "<Card>") {
			t.Errorf("hint without a suggestion: %s", d.Hint)
		}
		if d.Line == 8 && !strings.Contains(d.Hint, `type="warning"`) {
			t.Errorf("value hint without alias: %s", d.Hint)
		}
	}
}

func TestMismatchedCloseRecovers(t *testing.T) {
	r := run(t, "<Card title=\"a\">\n<Note>\ntext\n</Card>\n\nafter\n", "docs")
	r.code(t, diag.CodeComponentUnclosed, 2) // the Note
	for _, d := range r.diags {
		if d.Code == diag.CodeComponentUnclosed && d.Line == 1 {
			t.Errorf("the Card was closed and should not be reported as open: %v", r.diags)
		}
	}
	r.has(t, "<p>after</p>")
}

func TestCompatibilityAliases(t *testing.T) {
	r := run(t, "<CardGroup cols=\"3\">\n<Card title=\"A\" icon=\"bolt-lightning\">a</Card>\n</CardGroup>\n\n<AccordionGroup>\n<Accordion title=\"Q\" open=\"true\">A</Accordion>\n</AccordionGroup>\n\n<Check>ok</Check>\n", "docs")
	r.clean(t)
	r.has(t, "wm-columns--3", "wm-accordion-group", "<details class=\"wm-accordion\" open>", "wm-callout--success")
}

func TestRawHTMLIsForbidden(t *testing.T) {
	r := run(t, "Broken<br>line and <Card title=\"x\"> in the middle.\n\n<div>x</div>\n\n<!-- a comment -->\n", "docs")
	r.code(t, diag.CodeHTMLForbidden, 1)
	r.code(t, diag.CodeHTMLForbidden, 3)
	if n := slices.Index(r.codes, diag.CodeHTMLForbidden); n < 0 {
		t.Fatal("no html.forbidden")
	}
	r.lacks(t, "<div>x</div>", "<br>", "a comment")
	count := 0
	for _, c := range r.codes {
		if c == diag.CodeHTMLForbidden {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 HTML errors (br, inline Card, div), got %d: %v", count, r.diags)
	}
}

func TestIndentedCodeIsReportedAndContentKept(t *testing.T) {
	r := run(t, "# T\n\nText.\n\n    code\n\n[q](./missing.md)\n\n## Two\n", "docs")
	r.code(t, diag.CodeIndentedCode, 5)
	if len(r.a.Headings) != 1 || r.a.Headings[0].ID != "two" {
		t.Errorf("headings after the indented block were lost: %+v", r.a.Headings)
	}
	if len(r.a.Refs) != 1 {
		t.Errorf("links after the indented block were lost: %+v", r.a.Refs)
	}
}

func TestInlineComponents(t *testing.T) {
	r := run(t, "## Export <Badge color=\"clay\">New</Badge>\n\n<Badge>Beta</Badge> at the start of the line and <Icon name=\"rocket\" /> in the middle.\n\n| a | b |\n|---|---|\n| <Badge color=\"olive\">ok</Badge> | x |\n", "docs")
	r.clean(t)
	r.has(t, `<span class="wm-badge wm-badge--clay">New</span>`, `<span class="wm-badge wm-badge--gray">Beta</span>`, `wm-inline-icon`, `wm-badge--olive`)
	if r.a.Headings[0].ID != "export" {
		t.Errorf("heading id: %q", r.a.Headings[0].ID)
	}
	r2 := run(t, "Text <Badge>never closed\n", "docs")
	r2.code(t, diag.CodeHTMLForbidden, 1)
	if !strings.Contains(r2.diags[0].Hint, "same line") {
		t.Errorf("hint: %s", r2.diags[0].Hint)
	}
}

func TestTitleRules(t *testing.T) {
	r := run(t, "# Title\n\nText.\n\n# Another\n", "docs")
	if r.a.Title != "Title" {
		t.Errorf("title: %q", r.a.Title)
	}
	r.code(t, diag.CodeTitleMultipleH1, 5)
	r.lacks(t, ">Title<") // the first H1 moves to the layout

	d := &diag.List{}
	doc := Parse([]byte("# Body\n"), 3, d)
	Analyze(doc, "docs", true, d)
	if len(d.Items) != 1 || d.Items[0].Code != diag.CodeTitleDuplicate || d.Items[0].Line != 4 {
		t.Errorf("duplicate title with the frontmatter offset: %v", d.Items)
	}
}

func TestSlugs(t *testing.T) {
	for in, want := range map[string]string{
		"Advanced Setup":   "advanced-setup",
		"Why `Go`?":        "why-go",
		"Café crème & thé": "cafe-creme-the",
		"日本語":              "日本語",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWarnings(t *testing.T) {
	r := run(t, "Text with a note[^x].\n\n![](./a.png)\n\n```mermaid\ngraph TD\n```\n", "docs")
	r.code(t, diag.CodeLinkBroken, 1)
	r.code(t, diag.CodeAltMissing, 3)
	r.code(t, diag.CodeFenceLang, 5)
	r.clean(t)
}

func TestEscaping(t *testing.T) {
	r := run(t, `<Card title="&quot;><script>x</script>" href="javascript:alert(1)">a</Card>`+"\n\n[x](javascript:alert(1))\n", "docs")
	r.lacks(t, "<script>", `href="javascript`)
}

func TestHeroHighlight(t *testing.T) {
	r := run(t, "<Hero title=\"Docs that write themselves\" highlight=\"write themselves\">\nx\n</Hero>\n", "landing")
	r.clean(t)
	r.has(t, `<span class="wm-mark"><span class="wm-word">write</span> <span class="wm-word">themselves</span>`, "wm-mark__line")
	r2 := run(t, "<Hero title=\"Docs\" highlight=\"nothing\">\nx\n</Hero>\n", "landing")
	r2.code(t, diag.CodeAttrValue, 1)
}
