package components

import (
	"html"
	"sort"
	"strconv"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/icons"
	"github.com/yuin/goldmark/ast"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var calloutTypes = []string{"note", "tip", "info", "warning", "danger"}

var calloutIcons = map[string]string{
	"note": "notepad", "tip": "lightbulb", "info": "info", "warning": "alert", "danger": "octagon",
}

var registry = map[string]*Spec{}

// knownAliases maps names from other documentation systems (Mintlify,
// Docusaurus, MDX) to guidance, because agents often reach for them.
var knownAliases = map[string]string{
	"CardGroup":      "Use <Columns cols={2}> com <Card> dentro.",
	"Grid":           "Use <Columns cols={3}>.",
	"Image":          `Use <Figure src="..." alt="..." /> ou a sintaxe Markdown ![alt](src).`,
	"Img":            `Use <Figure src="..." alt="..." /> ou a sintaxe Markdown ![alt](src).`,
	"Frame":          `Use <Figure src="..." alt="..." caption="..." />.`,
	"Check":          "Use <Tip> ou <Callout type=\"tip\">.",
	"Caution":        "Use <Warning>.",
	"Important":      "Use <Warning>.",
	"Admonition":     `Use <Callout type="note">.`,
	"AccordionGroup": "Coloque vários <Accordion> em sequência; não há agrupador.",
	"Expandable":     `Use <Accordion title="...">.`,
	"Details":        `Use <Accordion title="...">.`,
	"TabItem":        `Use <Tab title="..."> dentro de <Tabs>.`,
	"CodeBlock":      "Use um bloco de código cercado (```). Para abas de código, <CodeGroup>.",
	"Link":           "Use a sintaxe Markdown [texto](destino) ou <Button>.",
	"Badge":          "Badge ainda não existe na v1.",
	"Icon":           "Ícones ainda não existem na v1.",
	"Section":        "Use títulos ## para seções.",
}

func register(s *Spec) { registry[s.Name] = s }

// Lookup returns the schema of a component, or nil.
func Lookup(name string) *Spec { return registry[name] }

// Names lists all component names in alphabetical order.
func Names() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// All returns every schema in documentation order: canonical components
// first, each followed by its aliases.
func All() []*Spec {
	var out []*Spec
	for _, name := range order {
		out = append(out, registry[name])
	}
	return out
}

var order []string

func canonical(name string) string {
	if s := registry[name]; s != nil && s.AliasOf != "" {
		return s.AliasOf
	}
	return name
}

func esc(s string) string { return html.EscapeString(s) }

func attr(w util.BufWriter, name, value string) {
	if (name == "href" || name == "src") && gmhtml.IsDangerousURL([]byte(value)) {
		value = ""
	}
	_, _ = w.WriteString(" " + name + `="` + esc(value) + `"`)
}

func init() {
	callout := &Spec{
		Name:    "Callout",
		Summary: "Bloco de destaque para notas, dicas e avisos.",
		Attrs: []AttrSpec{
			{Name: "type", Type: TEnum, Enum: calloutTypes, Default: "note", Doc: "Tom do destaque."},
			{Name: "title", Type: TString, Doc: "Título em negrito no topo do bloco."},
		},
		Example: "<Callout type=\"warning\" title=\"Antes de começar\">\nFaça backup do banco.\n</Callout>",
		Render:  renderCallout,
	}
	register(callout)
	order = append(order, "Callout")
	for _, t := range calloutTypes {
		name := strings.ToUpper(t[:1]) + t[1:]
		register(&Spec{
			Name:    name,
			Summary: "Atalho para <Callout type=\"" + t + "\">.",
			Attrs:   []AttrSpec{{Name: "title", Type: TString, Doc: "Título em negrito no topo do bloco."}},
			Preset:  map[string]any{"type": t},
			AliasOf: "Callout",
			Example: "<" + name + ">\nTexto em **Markdown**.\n</" + name + ">",
			Render:  renderCallout,
		})
		order = append(order, name)
	}

	register(&Spec{
		Name:    "Card",
		Summary: "Cartão com título e texto. Com href, o cartão inteiro vira link.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Título do cartão."},
			{Name: "href", Type: TURL, Doc: "Destino do link: rota, arquivo .md relativo ou URL https."},
			{Name: "cta", Type: TString, Doc: "Texto de chamada no rodapé do cartão (ex.: \"Ler guia\"). Exige href."},
		},
		Example: "<Card title=\"Instalação\" href=\"./instalacao.md\">\nColoque o binário no PATH.\n</Card>",
		Render:  renderCard,
	})
	order = append(order, "Card")

	register(&Spec{
		Name:    "Columns",
		Summary: "Grade responsiva. Em telas estreitas vira uma coluna.",
		Attrs: []AttrSpec{
			{Name: "cols", Type: TInt, Default: 2, Min: 1, Max: 4, Doc: "Número de colunas (1 a 4)."},
		},
		Example: "<Columns cols={2}>\n<Card title=\"A\">...</Card>\n<Card title=\"B\">...</Card>\n</Columns>",
		Render:  renderColumns,
	})
	order = append(order, "Columns")

	register(&Spec{
		Name:     "Tabs",
		Summary:  "Conteúdo alternativo em abas.",
		Children: ChildOnly,
		Allowed:  []string{"Tab"},
		Example:  "<Tabs>\n<Tab title=\"macOS\">\n...\n</Tab>\n<Tab title=\"Linux\">\n...\n</Tab>\n</Tabs>",
		Render:   renderTabs,
	})
	register(&Spec{
		Name:    "Tab",
		Summary: "Uma aba de <Tabs>.",
		Attrs:   []AttrSpec{{Name: "title", Type: TString, Required: true, Doc: "Rótulo da aba."}},
		Parent:  "Tabs",
		Example: "<Tab title=\"Linux\">\nConteúdo em Markdown.\n</Tab>",
		Render:  renderTab,
	})
	order = append(order, "Tabs", "Tab")

	register(&Spec{
		Name:     "CodeGroup",
		Summary:  "Vários blocos de código em abas. O rótulo vem de title=\"...\" na cerca ou da linguagem.",
		Children: ChildOnly,
		Allowed:  []string{ChildCode},
		Example:  "<CodeGroup>\n```bash title=\"npm\"\nnpm i x\n```\n```bash title=\"pnpm\"\npnpm add x\n```\n</CodeGroup>",
		Render:   renderCodeGroup,
	})
	order = append(order, "CodeGroup")

	register(&Spec{
		Name:     "Steps",
		Summary:  "Sequência numerada de passos.",
		Children: ChildOnly,
		Allowed:  []string{"Step"},
		Example:  "<Steps>\n<Step title=\"Instale\">\n...\n</Step>\n<Step title=\"Configure\">\n...\n</Step>\n</Steps>",
		Render:   renderSteps,
	})
	register(&Spec{
		Name:    "Step",
		Summary: "Um passo de <Steps>.",
		Attrs:   []AttrSpec{{Name: "title", Type: TString, Required: true, Doc: "Título do passo."}},
		Parent:  "Steps",
		Example: "<Step title=\"Instale\">\nRode o instalador.\n</Step>",
		Render:  renderStep,
	})
	order = append(order, "Steps", "Step")

	register(&Spec{
		Name:    "Accordion",
		Summary: "Seção recolhível.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Texto sempre visível."},
			{Name: "open", Type: TBool, Default: false, Doc: "Começa aberto."},
		},
		Example: "<Accordion title=\"Por que Go?\">\nBinário único.\n</Accordion>",
		Render:  renderAccordion,
	})
	order = append(order, "Accordion")

	register(&Spec{
		Name:     "Button",
		Summary:  "Link com aparência de botão. Botões consecutivos ficam lado a lado.",
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "href", Type: TURL, Required: true, Doc: "Destino."},
			{Name: "label", Type: TString, Required: true, Doc: "Texto do botão."},
			{Name: "variant", Type: TEnum, Enum: []string{"primary", "secondary"}, Default: "primary", Doc: "Estilo."},
		},
		Example: "<Button href=\"/guia\" label=\"Começar\" />",
		Render:  renderButton,
	})
	order = append(order, "Button")

	register(&Spec{
		Name:    "Hero",
		Summary: "Abertura de página de entrada, com título grande. Só no layout landing; o conteúdo vira o subtítulo e pode ter <Button>.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Título principal (vira o <h1> da página)."},
			{Name: "eyebrow", Type: TString, Doc: "Texto pequeno acima do título."},
			{Name: "align", Type: TEnum, Enum: []string{"center", "left"}, Default: "center", Doc: "Alinhamento."},
		},
		Layouts: []string{"landing"},
		Example: "<Hero title=\"Docs que se escrevem sozinhas\" eyebrow=\"Wavemark\">\nMarkdown vira site.\n\n<Button href=\"/guia\" label=\"Começar\" />\n</Hero>",
		Render:  renderHero,
	})
	order = append(order, "Hero")

	register(&Spec{
		Name:     "Figure",
		Summary:  "Imagem com legenda.",
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "src", Type: TAsset, Required: true, Doc: "Arquivo de imagem relativo à página, caminho absoluto no site ou URL https."},
			{Name: "alt", Type: TString, Required: true, Doc: "Descrição da imagem para leitores de tela."},
			{Name: "caption", Type: TString, Doc: "Legenda exibida abaixo."},
		},
		Example: "<Figure src=\"./arquitetura.png\" alt=\"Diagrama da arquitetura\" caption=\"Fluxo de uma requisição\" />",
		Render:  renderFigure,
	})
	order = append(order, "Figure")
}

func renderCallout(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div></aside>\n")
		return
	}
	t := n.Str("type")
	_, _ = w.WriteString(`<aside class="wm-callout wm-callout--` + esc(t) + `" role="note"`)
	attr(w, "aria-label", ctx.T(t))
	_, _ = w.WriteString(`><span class="wm-callout__icon">` + icons.SVG(calloutIcons[t]) + `</span><div class="wm-callout__body">`)
	if title := n.Str("title"); title != "" {
		_, _ = w.WriteString(`<p class="wm-callout__title">` + esc(title) + "</p>")
	}
	_, _ = w.WriteString("\n")
}

func renderCard(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	href := n.Str("href")
	if !entering {
		_, _ = w.WriteString("</div>")
		if cta := n.Str("cta"); cta != "" {
			_, _ = w.WriteString(`<span class="wm-card__cta">` + esc(cta) + icons.SVG("arrow-right") + "</span>")
		}
		if href != "" {
			_, _ = w.WriteString("</a>\n")
		} else {
			_, _ = w.WriteString("</div>\n")
		}
		return
	}
	if href != "" {
		_, _ = w.WriteString(`<a class="wm-card wm-card--link"`)
		attr(w, "href", ctx.URL(href))
		_, _ = w.WriteString(">")
	} else {
		_, _ = w.WriteString(`<div class="wm-card">`)
	}
	_, _ = w.WriteString(`<p class="wm-card__title">` + esc(n.Str("title")) + `</p><div class="wm-card__body">` + "\n")
}

func renderColumns(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div>\n")
		return
	}
	cols := n.Int("cols")
	if cols < 1 || cols > 4 {
		cols = 2
	}
	_, _ = w.WriteString(`<div class="wm-columns wm-columns--` + strconv.Itoa(cols) + `">` + "\n")
}

// assignPanels gives each child of a tab container a tab id and panel id.
func assignPanels(n *Node, ctx *RenderContext) []ast.Node {
	base := ctx.NextID("wm-tabs")
	var kids []ast.Node
	i := 0
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		ctx.Panels[c] = PanelID{Tab: base + "-tab-" + strconv.Itoa(i), Panel: base + "-panel-" + strconv.Itoa(i), Index: i}
		kids = append(kids, c)
		i++
	}
	return kids
}

func writeTabList(w util.BufWriter, labels []string, ids []PanelID) {
	_, _ = w.WriteString(`<div class="wm-tabs__list" role="tablist">`)
	for i, label := range labels {
		sel := "false"
		tabindex := "-1"
		if i == 0 {
			sel, tabindex = "true", "0"
		}
		_, _ = w.WriteString(`<button type="button" class="wm-tabs__tab" role="tab"`)
		attr(w, "id", ids[i].Tab)
		attr(w, "aria-controls", ids[i].Panel)
		attr(w, "aria-selected", sel)
		attr(w, "tabindex", tabindex)
		_, _ = w.WriteString(">" + esc(label) + "</button>")
	}
	_, _ = w.WriteString("</div>")
}

func renderTabs(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div>\n")
		return
	}
	kids := assignPanels(n, ctx)
	var labels []string
	var ids []PanelID
	for _, k := range kids {
		if tab, ok := k.(*Node); ok {
			labels = append(labels, tab.Str("title"))
			ids = append(ids, ctx.Panels[k])
		}
	}
	_, _ = w.WriteString(`<div class="wm-tabs" data-wm-tabs>`)
	writeTabList(w, labels, ids)
	_, _ = w.WriteString("\n")
}

func renderTab(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div>\n")
		return
	}
	id, ok := ctx.Panels[n]
	if !ok {
		id = PanelID{Tab: ctx.NextID("wm-tab"), Panel: ctx.NextID("wm-panel")}
	}
	_, _ = w.WriteString(`<div class="wm-tabs__panel" role="tabpanel"`)
	attr(w, "id", id.Panel)
	attr(w, "aria-labelledby", id.Tab)
	_, _ = w.WriteString(`><p class="wm-tabs__label">` + esc(n.Str("title")) + "</p>\n")
}

func renderCodeGroup(w util.BufWriter, source []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div>\n")
		return
	}
	kids := assignPanels(n, ctx)
	var labels []string
	var ids []PanelID
	for _, k := range kids {
		if fc, ok := k.(*ast.FencedCodeBlock); ok {
			label := Fence(fc, source).Label()
			if label == "" {
				label = "Code"
			}
			labels = append(labels, label)
			ids = append(ids, ctx.Panels[k])
		}
	}
	_, _ = w.WriteString(`<div class="wm-code wm-codegroup" data-wm-tabs><div class="wm-code__header">`)
	writeTabList(w, labels, ids)
	writeCopyButton(w, ctx)
	_, _ = w.WriteString("</div>\n")
}

// writeCopyButton writes the copy-to-clipboard button of a code block.
func writeCopyButton(w util.BufWriter, ctx *RenderContext) {
	_, _ = w.WriteString(`<button type="button" class="wm-code__copy" data-wm-copy`)
	attr(w, "aria-label", ctx.T("copyCode"))
	attr(w, "title", ctx.T("copy"))
	attr(w, "data-copied", ctx.T("copied"))
	_, _ = w.WriteString(">" + icons.SVG("copy") + icons.SVG("check") + "</button>")
}

// WriteCopyButton is exported for the code block renderer.
func WriteCopyButton(w util.BufWriter, ctx *RenderContext) { writeCopyButton(w, ctx) }

func renderSteps(w util.BufWriter, _ []byte, _ *Node, entering bool, _ *RenderContext) {
	if entering {
		_, _ = w.WriteString(`<ol class="wm-steps">` + "\n")
	} else {
		_, _ = w.WriteString("</ol>\n")
	}
}

func renderStep(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if entering {
		_, _ = w.WriteString(`<li class="wm-step"><p class="wm-step__title">` + esc(n.Str("title")) + `</p><div class="wm-step__body">` + "\n")
	} else {
		_, _ = w.WriteString("</div></li>\n")
	}
}

func renderAccordion(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div></details>\n")
		return
	}
	_, _ = w.WriteString(`<details class="wm-accordion"`)
	if n.Bool("open") {
		_, _ = w.WriteString(" open")
	}
	_, _ = w.WriteString(`><summary class="wm-accordion__summary"><span>` + esc(n.Str("title")) + "</span>" + icons.SVG("chevron-down") + `</summary><div class="wm-accordion__body">` + "\n")
}

func isButton(node ast.Node) bool {
	c, ok := node.(*Node)
	return ok && c.Name == "Button"
}

func renderButton(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		return
	}
	if !isButton(n.PreviousSibling()) {
		_, _ = w.WriteString(`<div class="wm-actions">`)
	}
	_, _ = w.WriteString(`<a class="wm-button wm-button--` + esc(n.Str("variant")) + `"`)
	attr(w, "href", ctx.URL(n.Str("href")))
	_, _ = w.WriteString(">" + esc(n.Str("label")) + "</a>")
	if !isButton(n.NextSibling()) {
		_, _ = w.WriteString("</div>\n")
	}
}

func renderHero(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div></section>\n")
		return
	}
	_, _ = w.WriteString(`<section class="wm-hero wm-hero--` + esc(n.Str("align")) + `">`)
	if e := n.Str("eyebrow"); e != "" {
		_, _ = w.WriteString(`<p class="wm-hero__eyebrow">` + esc(e) + "</p>")
	}
	_, _ = w.WriteString(`<h1 class="wm-hero__title">` + esc(n.Str("title")) + `</h1><div class="wm-hero__body">` + "\n")
}

func renderFigure(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		return
	}
	_, _ = w.WriteString(`<figure class="wm-figure"><img`)
	attr(w, "src", ctx.URL(n.Str("src")))
	attr(w, "alt", n.Str("alt"))
	_, _ = w.WriteString(` loading="lazy" decoding="async">`)
	if c := n.Str("caption"); c != "" {
		_, _ = w.WriteString("<figcaption>" + esc(c) + "</figcaption>")
	}
	_, _ = w.WriteString("</figure>\n")
}
