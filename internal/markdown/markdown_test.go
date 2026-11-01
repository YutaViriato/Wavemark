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
	html, err := Render(doc, RenderOptions{Layout: layout, Lang: "pt-BR"})
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
			t.Errorf("HTML sem %q:\n%s", s, r.html)
		}
	}
}

func (r result) lacks(t *testing.T, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if strings.Contains(r.html, s) {
			t.Errorf("HTML não deveria ter %q:\n%s", s, r.html)
		}
	}
}

func (r result) clean(t *testing.T) {
	t.Helper()
	for _, d := range r.diags {
		if d.Severity == diag.Error {
			t.Errorf("erro inesperado: %s", d)
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
	t.Errorf("esperava %s na linha %d; diagnósticos: %v", code, line, r.diags)
}

func TestNestedComponentsAndIndentation(t *testing.T) {
	r := run(t, `<Tabs>
  <Tab title="A">
    Texto **forte**.

    - item
      - aninhado
  </Tab>
  <Tab title="B">Uma linha</Tab>
</Tabs>
`, "docs")
	r.clean(t)
	r.has(t, `role="tab"`, `<strong>forte</strong>`, `<li>aninhado</li>`, `<p>Uma linha</p>`)
	if strings.Count(r.html, `role="tabpanel"`) != 2 {
		t.Errorf("esperava 2 painéis:\n%s", r.html)
	}
}

func TestSameNameNesting(t *testing.T) {
	r := run(t, "<Columns>\n<Columns cols={1}>\n<Card title=\"in\">x</Card>\n</Columns>\n<Card title=\"out\">y</Card>\n</Columns>\n", "docs")
	r.clean(t)
	if strings.Count(r.html, `wm-columns--`) != 2 {
		t.Errorf("esperava dois Columns:\n%s", r.html)
	}
	// "out" must be inside the outer grid, after the inner one closed.
	if i, j := strings.Index(r.html, "wm-columns--1"), strings.Index(r.html, ">out<"); i < 0 || j < i {
		t.Errorf("ordem errada:\n%s", r.html)
	}
}

func TestClosingTagInsideFenceIsCode(t *testing.T) {
	r := run(t, "<Tabs>\n<Tab title=\"x\">\n```html\n</Tab>\n```\n</Tab>\n</Tabs>\n", "docs")
	r.clean(t)
	r.has(t, `<span class="hl-nt">Tab</span>`, `role="tabpanel"`)
}

func TestSelfClosingInsideContainer(t *testing.T) {
	r := run(t, "<Hero title=\"Oi\">\nSub.\n<Button href=\"/\" label=\"Ir\" />\n<Button href=\"/b\">Outro</Button>\n</Hero>\n", "landing")
	r.clean(t)
	r.has(t, `<div class="wm-actions"><a class="wm-button wm-button--primary" href="/">Ir</a><a class="wm-button wm-button--primary" href="/b">Outro</a></div>`)
	r.lacks(t, "<p>Outro</p>")
}

func TestMultilineOpeningTag(t *testing.T) {
	r := run(t, "<Card\n  title=\"Longo\"\n  icon=\"rocket\"\n>\nCorpo.\n</Card>\n", "docs")
	r.clean(t)
	r.has(t, ">Longo<", "wm-sketch", "<p>Corpo.</p>")
}

func TestComponentErrors(t *testing.T) {
	r := run(t, `<Note>
sem fechar

</Steps>

<Crad title="x">a</Crad>

<Callout type="caution">x</Callout>

<Card title={x + 1}>a</Card>

<Card titl="a">b</Card>

<Accordion title="x" open="sim">y</Accordion>
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
			t.Errorf("dica sem sugestão: %s", d.Hint)
		}
		if d.Line == 8 && !strings.Contains(d.Hint, `type="warning"`) {
			t.Errorf("dica de valor sem alias: %s", d.Hint)
		}
	}
}

func TestMismatchedCloseRecovers(t *testing.T) {
	r := run(t, "<Card title=\"a\">\n<Note>\ntexto\n</Card>\n\ndepois\n", "docs")
	r.code(t, diag.CodeComponentUnclosed, 2) // the Note
	for _, d := range r.diags {
		if d.Code == diag.CodeComponentUnclosed && d.Line == 1 {
			t.Errorf("o Card foi fechado e não deveria aparecer como aberto: %v", r.diags)
		}
	}
	r.has(t, "<p>depois</p>")
}

func TestCompatibilityAliases(t *testing.T) {
	r := run(t, "<CardGroup cols=\"3\">\n<Card title=\"A\" icon=\"bolt-lightning\">a</Card>\n</CardGroup>\n\n<AccordionGroup>\n<Accordion title=\"P\" open=\"true\">R</Accordion>\n</AccordionGroup>\n\n<Check>ok</Check>\n", "docs")
	r.clean(t)
	r.has(t, "wm-columns--3", "wm-accordion-group", "<details class=\"wm-accordion\" open>", "wm-callout--success")
}

func TestRawHTMLIsForbidden(t *testing.T) {
	r := run(t, "Linha<br>quebrada e <Card title=\"x\"> no meio.\n\n<div>x</div>\n\n<!-- comentário -->\n", "docs")
	r.code(t, diag.CodeHTMLForbidden, 1)
	r.code(t, diag.CodeHTMLForbidden, 3)
	if n := slices.Index(r.codes, diag.CodeHTMLForbidden); n < 0 {
		t.Fatal("sem html.forbidden")
	}
	r.lacks(t, "<div>x</div>", "<br>", "comentário")
	count := 0
	for _, c := range r.codes {
		if c == diag.CodeHTMLForbidden {
			count++
		}
	}
	if count != 3 {
		t.Errorf("esperava 3 erros de HTML (br, Card inline, div), veio %d: %v", count, r.diags)
	}
}

func TestIndentedCodeIsReportedAndContentKept(t *testing.T) {
	r := run(t, "# T\n\nTexto.\n\n    codigo\n\n[q](./nao.md)\n\n## Dois\n", "docs")
	r.code(t, diag.CodeIndentedCode, 5)
	if len(r.a.Headings) != 1 || r.a.Headings[0].ID != "dois" {
		t.Errorf("títulos depois do bloco indentado se perderam: %+v", r.a.Headings)
	}
	if len(r.a.Refs) != 1 {
		t.Errorf("links depois do bloco indentado se perderam: %+v", r.a.Refs)
	}
}

func TestInlineComponents(t *testing.T) {
	r := run(t, "## Exportar <Badge color=\"clay\">Novo</Badge>\n\n<Badge>Beta</Badge> no começo da linha e <Icon name=\"rocket\" /> no meio.\n\n| a | b |\n|---|---|\n| <Badge color=\"olive\">ok</Badge> | x |\n", "docs")
	r.clean(t)
	r.has(t, `<span class="wm-badge wm-badge--clay">Novo</span>`, `<span class="wm-badge wm-badge--gray">Beta</span>`, `wm-inline-icon`, `wm-badge--olive`)
	if r.a.Headings[0].ID != "exportar" {
		t.Errorf("id do título: %q", r.a.Headings[0].ID)
	}
	r2 := run(t, "Texto <Badge>sem fechar\n", "docs")
	r2.code(t, diag.CodeHTMLForbidden, 1)
	if !strings.Contains(r2.diags[0].Hint, "mesma linha") {
		t.Errorf("dica: %s", r2.diags[0].Hint)
	}
}

func TestTitleRules(t *testing.T) {
	r := run(t, "# Título\n\nTexto.\n\n# Outro\n", "docs")
	if r.a.Title != "Título" {
		t.Errorf("título: %q", r.a.Title)
	}
	r.code(t, diag.CodeTitleMultipleH1, 5)
	r.lacks(t, ">Título<") // the first H1 moves to the layout

	d := &diag.List{}
	doc := Parse([]byte("# Corpo\n"), 3, d)
	Analyze(doc, "docs", true, d)
	if len(d.Items) != 1 || d.Items[0].Code != diag.CodeTitleDuplicate || d.Items[0].Line != 4 {
		t.Errorf("título duplicado com deslocamento do frontmatter: %v", d.Items)
	}
}

func TestSlugs(t *testing.T) {
	for in, want := range map[string]string{
		"Configuração avançada": "configuracao-avancada",
		"Por que `Go`?":         "por-que-go",
		"Ação & reação":         "acao-reacao",
		"日本語":                   "日本語",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, quero %q", in, got, want)
		}
	}
}

func TestWarnings(t *testing.T) {
	r := run(t, "Texto com nota[^x].\n\n![](./a.png)\n\n```mermaid\ngraph TD\n```\n", "docs")
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
	r := run(t, "<Hero title=\"Docs que se escrevem\" highlight=\"se escrevem\">\nx\n</Hero>\n", "landing")
	r.clean(t)
	r.has(t, `<span class="wm-mark"><span class="wm-word">se</span> <span class="wm-word">escrevem</span>`, "wm-mark__line")
	r2 := run(t, "<Hero title=\"Docs\" highlight=\"nada\">\nx\n</Hero>\n", "landing")
	r2.code(t, diag.CodeAttrValue, 1)
}
