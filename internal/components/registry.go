package components

import (
	"html"
	"sort"
	"strconv"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/icons"
	"github.com/YutaViriato/Wavemark/internal/sketch"
	"github.com/yuin/goldmark/ast"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var calloutTypes = []string{"note", "tip", "info", "success", "warning", "danger"}

var calloutIcons = map[string]string{
	"note": "notepad", "tip": "lightbulb", "info": "info", "success": "check-circle", "warning": "alert", "danger": "octagon",
}

var badgeColors = []string{"gray", "clay", "olive", "sky", "fig"}

var registry = map[string]*Spec{}

// knownAliases maps names from other documentation systems to guidance,
// for names that have no direct equivalent.
var knownAliases = map[string]string{
	"Grid":        "Use <Columns cols={3}>.",
	"Image":       `Use <Figure src="..." alt="..." /> ou a sintaxe Markdown ![alt](src).`,
	"Img":         `Use <Figure src="..." alt="..." /> ou a sintaxe Markdown ![alt](src).`,
	"Frame":       `Use <Figure src="..." alt="..." caption="..." />.`,
	"Caution":     "Use <Warning>.",
	"Important":   "Use <Warning>.",
	"Admonition":  `Use <Callout type="note">.`,
	"Expandable":  `Use <Accordion title="...">.`,
	"Details":     `Use <Accordion title="...">.`,
	"TabItem":     `Use <Tab title="..."> dentro de <Tabs>.`,
	"CodeBlock":   "Use um bloco de código cercado (```). Para abas de código, <CodeGroup>.",
	"Link":        "Use a sintaxe Markdown [texto](destino) ou <Button>.",
	"Blockquote":  `Use '> texto' ou <Quote author="...">.`,
	"Testimonial": `Use <Quote author="..." role="...">.`,
	"Pricing":     "Use uma tabela Markdown ou <Columns> com <Card>.",
	"Banner":      `Use <Section tone="oat"> no layout landing ou <Callout> nas páginas docs.`,
	"Kbd":         "Use código inline: `Ctrl+K`.",
	"Mermaid":     "Diagramas ainda não são renderizados. Use uma imagem SVG com <Figure>.",
	"Divider":     "Use uma linha com --- para separar seções.",
	"Spacer":      "O espaçamento é automático; remova o espaçador.",
}

// attrAliases explains attributes other systems use, keyed "Component.attr".
var attrAliases = map[string]string{
	"Hero.subtitle":         "O subtítulo é o conteúdo entre <Hero> e </Hero>.",
	"Hero.description":      "O subtítulo é o conteúdo entre <Hero> e </Hero>.",
	"Hero.tagline":          "Use eyebrow para o texto pequeno acima do título; o subtítulo é o conteúdo.",
	"Hero.img":              "Use image.",
	"Hero.src":              "Use image.",
	"Card.description":      "O texto do cartão é o conteúdo entre <Card> e </Card>.",
	"Card.image":            `Use icon="nome" para um ícone desenhado; imagens vão no conteúdo com ![alt](src).`,
	"Card.img":              `Use icon="nome" para um ícone desenhado; imagens vão no conteúdo com ![alt](src).`,
	"Card.link":             "Use href.",
	"Card.url":              "Use href.",
	"Card.color":            "Cartões seguem o tema; não há cor por cartão.",
	"Card.horizontal":       "Não há variação horizontal.",
	"Button.text":           "Use label, ou escreva o texto entre as tags.",
	"Button.title":          "Use label, ou escreva o texto entre as tags.",
	"Button.link":           "Use href.",
	"Button.url":            "Use href.",
	"Button.size":           "O tamanho segue o layout (maior em landing).",
	"Button.icon":           "Botões não têm ícone.",
	"Figure.title":          "Use caption.",
	"Figure.image":          "Use src.",
	"Tab.label":             "Use title.",
	"Tab.name":              "Use title.",
	"Step.label":            "Use title.",
	"Accordion.label":       "Use title.",
	"Accordion.defaultOpen": "Use open.",
	"Callout.icon":          "O ícone vem do type.",
	"Quote.cite":            "Use author e role.",
	"Quote.name":            "Use author.",
	"Columns.gap":           "O espaçamento é automático.",
}

// valueAliases explains enum values from other systems, keyed
// "Component.attr=value".
var valueAliases = map[string]string{
	"Button.variant=outline": `Use variant="secondary".`,
	"Button.variant=ghost":   `Use variant="secondary".`,
	"Button.variant=link":    "Use um link Markdown [texto](destino).",
	"Button.variant=default": `Use variant="primary".`,
	"Callout.type=error":     `Use type="danger".`,
	"Callout.type=caution":   `Use type="warning".`,
	"Callout.type=important": `Use type="warning".`,
	"Callout.type=check":     `Use type="success".`,
	"Callout.type=hint":      `Use type="tip".`,
	"Badge.color=red":        `Use color="clay".`,
	"Badge.color=green":      `Use color="olive".`,
	"Badge.color=blue":       `Use color="sky".`,
	"Badge.color=purple":     `Use color="fig".`,
	"Badge.color=orange":     `Use color="clay".`,
	"Badge.color=yellow":     `Use color="clay".`,
	"Section.tone=gray":      `Use tone="muted".`,
	"Section.tone=light":     `Use tone="muted".`,
	"Section.tone=black":     `Use tone="dark".`,
	"Section.tone=accent":    `Use tone="clay".`,
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

// All returns every schema in documentation order.
func All() []*Spec {
	var out []*Spec
	for _, name := range order {
		out = append(out, registry[name])
	}
	return out
}

var order []string

func add(s *Spec) {
	register(s)
	order = append(order, s.Name)
}

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
	add(callout)
	for _, alias := range []struct{ name, t string }{{"Note", "note"}, {"Tip", "tip"}, {"Info", "info"}, {"Success", "success"}, {"Check", "success"}, {"Warning", "warning"}, {"Danger", "danger"}} {
		add(&Spec{
			Name:    alias.name,
			Summary: "Atalho para <Callout type=\"" + alias.t + "\">.",
			Attrs:   []AttrSpec{{Name: "title", Type: TString, Doc: "Título em negrito no topo do bloco."}},
			Preset:  map[string]any{"type": alias.t},
			AliasOf: "Callout",
			Example: "<" + alias.name + ">\nTexto em **Markdown**.\n</" + alias.name + ">",
			Render:  renderCallout,
		})
	}

	add(&Spec{
		Name:    "Card",
		Summary: "Cartão com ícone opcional, título e texto. Com href, o cartão inteiro vira link.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Título do cartão."},
			{Name: "href", Type: TURL, Doc: "Destino do link: rota, arquivo .md relativo ou URL https."},
			{Name: "icon", Type: TIcon, Doc: "Ícone desenhado à mão acima do título (veja a lista de ícones)."},
			{Name: "cta", Type: TString, Doc: "Texto de chamada no rodapé do cartão (ex.: \"Ler guia\"). Exige href."},
		},
		Example: "<Card title=\"Instalação\" icon=\"rocket\" href=\"./instalacao.md\">\nColoque o binário no PATH.\n</Card>",
		Render:  renderCard,
	})

	columns := &Spec{
		Name:    "Columns",
		Summary: "Grade responsiva. Em telas estreitas vira uma coluna.",
		Attrs: []AttrSpec{
			{Name: "cols", Type: TInt, Default: 2, Min: 1, Max: 4, Doc: "Número de colunas (1 a 4)."},
		},
		Example: "<Columns cols={2}>\n<Card title=\"A\">...</Card>\n<Card title=\"B\">...</Card>\n</Columns>",
		Render:  renderColumns,
	}
	add(columns)
	add(&Spec{Name: "CardGroup", Summary: "Atalho para <Columns> (nome usado pelo Mintlify).", Attrs: columns.Attrs, AliasOf: "Columns", Example: "<CardGroup cols={2}>\n<Card title=\"A\">...</Card>\n</CardGroup>", Render: renderColumns})

	add(&Spec{
		Name:     "Tabs",
		Summary:  "Conteúdo alternativo em abas.",
		Children: ChildOnly,
		Allowed:  []string{"Tab"},
		Example:  "<Tabs>\n<Tab title=\"macOS\">\n...\n</Tab>\n<Tab title=\"Linux\">\n...\n</Tab>\n</Tabs>",
		Render:   renderTabs,
	})
	add(&Spec{
		Name:    "Tab",
		Summary: "Uma aba de <Tabs>.",
		Attrs:   []AttrSpec{{Name: "title", Type: TString, Required: true, Doc: "Rótulo da aba."}},
		Parent:  "Tabs",
		Example: "<Tab title=\"Linux\">\nConteúdo em Markdown.\n</Tab>",
		Render:  renderTab,
	})

	add(&Spec{
		Name:     "CodeGroup",
		Summary:  "Vários blocos de código em abas. O rótulo vem de title=\"...\" na cerca ou da linguagem.",
		Children: ChildOnly,
		Allowed:  []string{ChildCode},
		Example:  "<CodeGroup>\n```bash title=\"npm\"\nnpm i x\n```\n```bash title=\"pnpm\"\npnpm add x\n```\n</CodeGroup>",
		Render:   renderCodeGroup,
	})

	add(&Spec{
		Name:     "Steps",
		Summary:  "Sequência numerada de passos, com marcadores desenhados à mão.",
		Children: ChildOnly,
		Allowed:  []string{"Step"},
		Example:  "<Steps>\n<Step title=\"Instale\">\n...\n</Step>\n<Step title=\"Configure\">\n...\n</Step>\n</Steps>",
		Render:   renderSteps,
	})
	add(&Spec{
		Name:    "Step",
		Summary: "Um passo de <Steps>.",
		Attrs:   []AttrSpec{{Name: "title", Type: TString, Required: true, Doc: "Título do passo."}},
		Parent:  "Steps",
		Example: "<Step title=\"Instale\">\nRode o instalador.\n</Step>",
		Render:  renderStep,
	})

	add(&Spec{
		Name:    "Accordion",
		Summary: "Seção recolhível.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Texto sempre visível."},
			{Name: "open", Type: TBool, Default: false, Doc: "Começa aberto."},
		},
		Example: "<Accordion title=\"Por que Go?\">\nBinário único.\n</Accordion>",
		Render:  renderAccordion,
	})
	add(&Spec{
		Name:     "AccordionGroup",
		Summary:  "Agrupa vários <Accordion> em um bloco contínuo, como uma lista de perguntas frequentes.",
		Children: ChildOnly,
		Allowed:  []string{"Accordion"},
		Example:  "<AccordionGroup>\n<Accordion title=\"Pergunta 1\">\nResposta.\n</Accordion>\n<Accordion title=\"Pergunta 2\">\nResposta.\n</Accordion>\n</AccordionGroup>",
		Render:   renderAccordionGroup,
	})

	add(&Spec{
		Name:    "Button",
		Summary: "Link com aparência de botão. O texto vem de label ou do conteúdo. Botões consecutivos ficam lado a lado.",
		Attrs: []AttrSpec{
			{Name: "href", Type: TURL, Required: true, Doc: "Destino."},
			{Name: "label", Type: TString, Doc: "Texto do botão (ou escreva o texto entre as tags)."},
			{Name: "variant", Type: TEnum, Enum: []string{"primary", "secondary"}, Default: "primary", Doc: "Estilo."},
		},
		Example:      "<Button href=\"/guia\" label=\"Começar\" />",
		Render:       renderButton,
		SkipChildren: true,
	})

	add(&Spec{
		Name:    "Hero",
		Summary: "Abertura de página de entrada, com título grande. Só no layout landing. O conteúdo vira o subtítulo e pode ter <Button>.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Título principal (vira o <h1> da página)."},
			{Name: "eyebrow", Type: TString, Doc: "Texto pequeno acima do título."},
			{Name: "highlight", Type: TString, Doc: "Trecho do título sublinhado com um traço desenhado à mão. Precisa aparecer exatamente no title."},
			{Name: "image", Type: TAsset, Doc: "Imagem ao lado do texto (SVG ou PNG). Com image, o hero fica em duas colunas."},
			{Name: "alt", Type: TString, Doc: "Descrição da imagem. Obrigatório quando há image."},
			{Name: "align", Type: TEnum, Enum: []string{"center", "left"}, Default: "center", Doc: "Alinhamento do texto (sem image)."},
		},
		Layouts: []string{"landing"},
		Example: "<Hero title=\"Docs que se escrevem sozinhas\" highlight=\"sozinhas\" eyebrow=\"Wavemark\">\nMarkdown vira site.\n\n<Button href=\"/guia\" label=\"Começar\" />\n</Hero>",
		Render:  renderHero,
	})

	add(&Spec{
		Name:    "Section",
		Summary: "Faixa de largura total com fundo próprio, para destacar um trecho da landing (chamada final, depoimentos, números).",
		Attrs: []AttrSpec{
			{Name: "tone", Type: TEnum, Enum: []string{"muted", "oat", "clay", "dark"}, Default: "muted", Doc: "Cor de fundo."},
			{Name: "align", Type: TEnum, Enum: []string{"left", "center"}, Default: "left", Doc: "Alinhamento do texto."},
		},
		Layouts: []string{"landing"},
		Example: "<Section tone=\"oat\" align=\"center\">\n## Pronto para começar?\n\n<Button href=\"/guia\" label=\"Ler o guia\" />\n</Section>",
		Render:  renderSection,
	})

	add(&Spec{
		Name:    "Quote",
		Summary: "Citação em destaque, com autor. Serve para depoimentos e epígrafes.",
		Attrs: []AttrSpec{
			{Name: "author", Type: TString, Doc: "Quem disse."},
			{Name: "role", Type: TString, Doc: "Cargo, organização ou fonte."},
		},
		Example: "<Quote author=\"Ada Lovelace\" role=\"Notas sobre a Máquina Analítica\">\nA máquina não tem pretensão de originar nada.\n</Quote>",
		Render:  renderQuote,
	})

	add(&Spec{
		Name:     "Stat",
		Summary:  "Número em destaque com legenda. Use vários dentro de <Columns>.",
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "value", Type: TString, Required: true, Doc: "O número, como texto (\"27\", \"180 mil\", \"99,9%\")."},
			{Name: "label", Type: TString, Required: true, Doc: "O que o número mede."},
		},
		Example: "<Stat value=\"27\" label=\"ensaios publicados\" />",
		Render:  renderStat,
	})

	add(&Spec{
		Name:     "PageList",
		Summary:  "Lista automática das páginas de uma pasta, com data, autor e descrição. Ideal para blogs e índices.",
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "dir", Type: TDir, Default: "./", Doc: "Pasta a listar, relativa à página (padrão: a pasta da própria página)."},
			{Name: "sort", Type: TEnum, Enum: []string{"auto", "date", "title", "nav"}, Default: "auto", Doc: "Ordem: date (mais recentes primeiro), title, nav (ordem da navegação) ou auto (date se as páginas tiverem data)."},
			{Name: "limit", Type: TInt, Default: 0, Min: 0, Max: 500, Doc: "Máximo de itens (0 = todos)."},
			{Name: "style", Type: TEnum, Enum: []string{"list", "cards"}, Default: "list", Doc: "Lista com linhas ou grade de cartões."},
		},
		Example: "<PageList dir=\"./blog\" sort=\"date\" limit={5} />",
		Render:  renderPageList,
	})

	add(&Spec{
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

	add(&Spec{
		Name:         "Badge",
		Summary:      "Etiqueta curta dentro do texto, de títulos ou de células de tabela (\"Novo\", \"Beta\", \"v2.1\").",
		Inline:       true,
		Children:     ChildText,
		Attrs:        []AttrSpec{{Name: "color", Type: TEnum, Enum: badgeColors, Default: "gray", Doc: "Cor."}},
		Example:      "<Badge color=\"clay\">Novo</Badge>",
		RenderInline: renderBadge,
	})
	add(&Spec{
		Name:     "Icon",
		Summary:  "Ícone desenhado à mão dentro do texto.",
		Inline:   true,
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "name", Type: TIcon, Required: true, Doc: "Nome do ícone."},
			{Name: "label", Type: TString, Doc: "Descrição para leitores de tela. Sem label, o ícone é decorativo."},
		},
		Example:      "<Icon name=\"sparkles\" />",
		RenderInline: renderIcon,
	})

	// Rules that need more than one attribute.
	registry["Card"].check = func(c *Core, diags *diag.List) {
		if c.Str("cta") != "" && c.Str("href") == "" {
			diags.Errorf(c.Line, c.Col, diag.CodeAttrValue, "Adicione href ou remova cta.", "'cta' em <Card> exige 'href'")
		}
	}
	registry["Hero"].check = func(c *Core, diags *diag.List) {
		if h := c.Str("highlight"); h != "" && !strings.Contains(c.Str("title"), h) {
			diags.Errorf(c.Line, c.Col, diag.CodeAttrValue, "Copie o trecho exatamente como está no title.", "highlight %q não aparece no title de <Hero>", h)
		}
		if c.Str("image") != "" && c.Str("alt") == "" {
			diags.Errorf(c.Line, c.Col, diag.CodeAttrMissing, `Descreva a imagem: alt="..."`, "<Hero image> exige o atributo 'alt'")
		}
	}
}

// Check runs component rules that involve several attributes.
func Check(c *Core, diags *diag.List) {
	if c.Spec != nil && c.Spec.check != nil {
		c.Spec.check(c, diags)
	}
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

func iconSize(ctx *RenderContext, docs, landing int) int {
	if ctx.Layout == "landing" {
		return landing
	}
	return docs
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
	if icon := n.Str("icon"); icon != "" {
		_, _ = w.WriteString(`<span class="wm-card__icon">` + sketch.Icon(icon, sketch.Options{Size: iconSize(ctx, 36, 48)}) + "</span>")
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
	_, _ = w.WriteString(`<span class="wm-tabs__indicator" aria-hidden="true"></span></div>`)
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

func renderStep(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div></li>\n")
		return
	}
	i := 1
	for p := n.PreviousSibling(); p != nil; p = p.PreviousSibling() {
		i++
	}
	_, _ = w.WriteString(`<li class="wm-step"><span class="wm-step__marker">` + sketch.Circle(uint64(i)*7919+uint64(len(n.Str("title"))), sketch.Options{Class: "wm-step__ring"}) + `<span class="wm-step__num">` + strconv.Itoa(i) + `</span></span><p class="wm-step__title">` + esc(n.Str("title")) + `</p><div class="wm-step__body">` + "\n")
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

func renderAccordionGroup(w util.BufWriter, _ []byte, _ *Node, entering bool, _ *RenderContext) {
	if entering {
		_, _ = w.WriteString(`<div class="wm-accordion-group">` + "\n")
	} else {
		_, _ = w.WriteString("</div>\n")
	}
}

func isButton(node ast.Node) bool {
	c, ok := node.(*Node)
	return ok && c.Name == "Button"
}

// TextOf returns the plain text of a node's descendants.
func TextOf(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Value(src))
			if t.SoftLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(strings.Fields(b.String()), " ")
}

func renderButton(w util.BufWriter, src []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		if !isButton(n.NextSibling()) {
			_, _ = w.WriteString("</div>\n")
		}
		return
	}
	if !isButton(n.PreviousSibling()) {
		_, _ = w.WriteString(`<div class="wm-actions">`)
	}
	label := n.Str("label")
	if label == "" {
		label = TextOf(n, src)
	}
	_, _ = w.WriteString(`<a class="wm-button wm-button--` + esc(n.Str("variant")) + `"`)
	attr(w, "href", ctx.URL(n.Str("href")))
	_, _ = w.WriteString(">" + esc(label) + "</a>")
}

func renderHero(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	img := n.Str("image")
	if !entering {
		_, _ = w.WriteString("</div></div>")
		if img != "" {
			_, _ = w.WriteString(`<div class="wm-hero__media"><img`)
			attr(w, "src", ctx.URL(img))
			attr(w, "alt", n.Str("alt"))
			_, _ = w.WriteString(` decoding="async"></div>`)
		}
		_, _ = w.WriteString("</section>\n")
		return
	}
	cls := "wm-hero wm-hero--" + n.Str("align")
	if img != "" {
		cls = "wm-hero wm-hero--split"
	}
	_, _ = w.WriteString(`<section class="` + cls + `"><div class="wm-hero__text">`)
	if e := n.Str("eyebrow"); e != "" {
		_, _ = w.WriteString(`<p class="wm-hero__eyebrow">` + esc(e) + "</p>")
	}
	_, _ = w.WriteString(`<h1 class="wm-hero__title">` + heroTitle(n.Str("title"), n.Str("highlight")) + `</h1><div class="wm-hero__body">` + "\n")
}

// heroTitle splits the title into words (animated one by one) and wraps the
// highlighted phrase with a hand-drawn underline.
func heroTitle(title, highlight string) string {
	words := func(s string) string {
		var b strings.Builder
		for i, word := range strings.Fields(s) {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(`<span class="wm-word">` + esc(word) + `</span>`)
		}
		return b.String()
	}
	if highlight == "" {
		return words(title)
	}
	i := strings.Index(title, highlight)
	if i < 0 {
		return words(title)
	}
	before, after := title[:i], title[i+len(highlight):]
	var b strings.Builder
	b.WriteString(words(before))
	if strings.HasSuffix(before, " ") {
		b.WriteString(" ")
	}
	b.WriteString(`<span class="wm-mark">` + words(highlight) + sketch.Underline(sketch.Seed(highlight), sketch.Options{Class: "wm-mark__line"}) + `</span>`)
	if strings.HasPrefix(after, " ") {
		b.WriteString(" ")
	}
	b.WriteString(words(after))
	return b.String()
}

func renderSection(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div></section>\n")
		return
	}
	_, _ = w.WriteString(`<section class="wm-section wm-section--` + esc(n.Str("tone")) + ` wm-section--` + esc(n.Str("align")) + `"><div class="wm-section__inner">` + "\n")
}

func renderQuote(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if entering {
		_, _ = w.WriteString(`<figure class="wm-quote"><blockquote class="wm-quote__text">` + "\n")
		return
	}
	_, _ = w.WriteString("</blockquote>")
	author, role := n.Str("author"), n.Str("role")
	if author != "" || role != "" {
		_, _ = w.WriteString(`<figcaption class="wm-quote__by">`)
		if author != "" {
			_, _ = w.WriteString(`<span class="wm-quote__author">` + esc(author) + "</span>")
		}
		if role != "" {
			_, _ = w.WriteString(`<span class="wm-quote__role">` + esc(role) + "</span>")
		}
		_, _ = w.WriteString("</figcaption>")
	}
	_, _ = w.WriteString("</figure>\n")
}

func renderStat(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if !entering {
		return
	}
	_, _ = w.WriteString(`<div class="wm-stat"><p class="wm-stat__value">` + esc(n.Str("value")) + `</p><p class="wm-stat__label">` + esc(n.Str("label")) + "</p></div>\n")
}

func renderPageList(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		return
	}
	pages := ctx.Pages(n.Str("dir"))
	sortBy := n.Str("sort")
	if sortBy == "auto" {
		sortBy = "nav"
		for _, p := range pages {
			if !p.Date.IsZero() {
				sortBy = "date"
				break
			}
		}
	}
	switch sortBy {
	case "date":
		sort.SliceStable(pages, func(i, j int) bool { return pages[i].Date.After(pages[j].Date) })
	case "title":
		sort.SliceStable(pages, func(i, j int) bool { return strings.ToLower(pages[i].Title) < strings.ToLower(pages[j].Title) })
	}
	if lim := n.Int("limit"); lim > 0 && len(pages) > lim {
		pages = pages[:lim]
	}
	if n.Str("style") == "cards" {
		_, _ = w.WriteString(`<div class="wm-columns wm-columns--3 wm-pagelist wm-pagelist--cards">`)
		for _, p := range pages {
			_, _ = w.WriteString(`<a class="wm-card wm-card--link"`)
			attr(w, "href", p.Route)
			_, _ = w.WriteString(">")
			if p.Icon != "" {
				_, _ = w.WriteString(`<span class="wm-card__icon">` + sketch.Icon(p.Icon, sketch.Options{Size: iconSize(ctx, 36, 48)}) + "</span>")
			}
			if meta := pageMeta(p, ctx); meta != "" {
				_, _ = w.WriteString(`<p class="wm-pagelist__meta">` + meta + "</p>")
			}
			_, _ = w.WriteString(`<p class="wm-card__title">` + esc(p.Title) + `</p><div class="wm-card__body">`)
			if p.Description != "" {
				_, _ = w.WriteString("<p>" + esc(p.Description) + "</p>")
			}
			_, _ = w.WriteString("</div></a>")
		}
		_, _ = w.WriteString("</div>\n")
		return
	}
	_, _ = w.WriteString(`<ul class="wm-pagelist">`)
	for _, p := range pages {
		_, _ = w.WriteString(`<li class="wm-pagelist__item"><a class="wm-pagelist__link"`)
		attr(w, "href", p.Route)
		_, _ = w.WriteString(">")
		if meta := pageMeta(p, ctx); meta != "" {
			_, _ = w.WriteString(`<span class="wm-pagelist__meta">` + meta + "</span>")
		}
		_, _ = w.WriteString(`<span class="wm-pagelist__title">` + esc(p.Title) + "</span>")
		if p.Description != "" {
			_, _ = w.WriteString(`<span class="wm-pagelist__desc">` + esc(p.Description) + "</span>")
		}
		_, _ = w.WriteString("</a></li>")
	}
	_, _ = w.WriteString("</ul>\n")
}

func pageMeta(p PageInfo, ctx *RenderContext) string {
	var parts []string
	if !p.Date.IsZero() {
		parts = append(parts, `<time datetime="`+p.Date.Format("2006-01-02")+`">`+esc(FormatDate(p.Date, ctx.Lang))+`</time>`)
	}
	if p.Author != "" {
		parts = append(parts, esc(p.Author))
	}
	return strings.Join(parts, `<span aria-hidden="true"> · </span>`)
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

func renderBadge(w util.BufWriter, n *InlineNode, _ *RenderContext) {
	_, _ = w.WriteString(`<span class="wm-badge wm-badge--` + esc(n.Str("color")) + `">` + esc(strings.TrimSpace(n.Content)) + "</span>")
}

func renderIcon(w util.BufWriter, n *InlineNode, _ *RenderContext) {
	_, _ = w.WriteString(`<span class="wm-inline-icon">` + sketch.Icon(n.Str("name"), sketch.Options{Label: n.Str("label")}) + "</span>")
}
