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
	"Image":       `Use <Figure src="..." alt="..." /> or Markdown image syntax ![alt](src).`,
	"Img":         `Use <Figure src="..." alt="..." /> or Markdown image syntax ![alt](src).`,
	"Frame":       `Use <Figure src="..." alt="..." caption="..." />.`,
	"Caution":     "Use <Warning>.",
	"Important":   "Use <Warning>.",
	"Admonition":  `Use <Callout type="note">.`,
	"Expandable":  `Use <Accordion title="...">.`,
	"Details":     `Use <Accordion title="...">.`,
	"TabItem":     `Use <Tab title="..."> inside <Tabs>.`,
	"CodeBlock":   "Use a fenced code block (```). For tabbed code, <CodeGroup>.",
	"Link":        "Use Markdown link syntax [text](destination) or <Button>.",
	"Blockquote":  `Use '> text' or <Quote author="...">.`,
	"Testimonial": `Use <Quote author="..." role="...">.`,
	"Pricing":     "Use a Markdown table, or <Columns> with <Card>.",
	"Banner":      `Use <Section tone="oat"> in the landing layout, or <Callout> on docs pages.`,
	"Kbd":         "Use inline code: `Ctrl+K`.",
	"Mermaid":     "Use <Flow> for step diagrams, or an SVG image with <Figure>.",
	"Divider":     "Use a line with --- to separate sections.",
	"Spacer":      "Spacing is automatic; remove the spacer.",
}

// attrAliases explains attributes other systems use, keyed "Component.attr".
var attrAliases = map[string]string{
	"Hero.subtitle":         "The subtitle is the content between <Hero> and </Hero>.",
	"Hero.description":      "The subtitle is the content between <Hero> and </Hero>.",
	"Hero.tagline":          "Use eyebrow for the small text above the title; the subtitle is the content.",
	"Hero.img":              "Use image.",
	"Hero.src":              "Use image.",
	"Card.description":      "The card text is the content between <Card> and </Card>.",
	"Card.image":            `Use icon="name" for a hand-drawn icon; images go in the content as ![alt](src).`,
	"Card.img":              `Use icon="name" for a hand-drawn icon; images go in the content as ![alt](src).`,
	"Card.link":             "Use href.",
	"Card.url":              "Use href.",
	"Card.color":            "Cards follow the theme; there is no per-card color.",
	"Card.horizontal":       "There is no horizontal variant.",
	"Button.text":           "Use label, or write the text between the tags.",
	"Button.title":          "Use label, or write the text between the tags.",
	"Button.link":           "Use href.",
	"Button.url":            "Use href.",
	"Button.size":           "The size follows the layout (larger on landing pages).",
	"Button.icon":           "Buttons have no icon.",
	"Figure.title":          "Use caption.",
	"Figure.image":          "Use src.",
	"Tab.label":             "Use title.",
	"Tab.name":              "Use title.",
	"Step.label":            "Use title.",
	"Accordion.label":       "Use title.",
	"Accordion.defaultOpen": "Use open.",
	"Callout.icon":          "The icon comes from type.",
	"Quote.cite":            "Use author and role.",
	"Quote.name":            "Use author.",
	"Columns.gap":           "Spacing is automatic.",
}

// valueAliases explains enum values from other systems, keyed
// "Component.attr=value".
var valueAliases = map[string]string{
	"Button.variant=outline": `Use variant="secondary".`,
	"Button.variant=ghost":   `Use variant="secondary".`,
	"Button.variant=link":    "Use a Markdown link [text](destination).",
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
		Summary: "Highlighted block for notes, tips and warnings.",
		Attrs: []AttrSpec{
			{Name: "type", Type: TEnum, Enum: calloutTypes, Default: "note", Doc: "Tone of the callout."},
			{Name: "title", Type: TString, Doc: "Bold title at the top of the block."},
		},
		Example: "<Callout type=\"warning\" title=\"Before you start\">\nBack up the database.\n</Callout>",
		Render:  renderCallout,
	}
	add(callout)
	for _, alias := range []struct{ name, t string }{{"Note", "note"}, {"Tip", "tip"}, {"Info", "info"}, {"Success", "success"}, {"Check", "success"}, {"Warning", "warning"}, {"Danger", "danger"}} {
		add(&Spec{
			Name:    alias.name,
			Summary: "Shorthand for <Callout type=\"" + alias.t + "\">.",
			Attrs:   []AttrSpec{{Name: "title", Type: TString, Doc: "Bold title at the top of the block."}},
			Preset:  map[string]any{"type": alias.t},
			AliasOf: "Callout",
			Example: "<" + alias.name + ">\nText in **Markdown**.\n</" + alias.name + ">",
			Render:  renderCallout,
		})
	}

	add(&Spec{
		Name:    "Card",
		Summary: "Card with an optional icon, a title and text. With href, the whole card is a link.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Card title."},
			{Name: "href", Type: TURL, Doc: "Link destination: a route, a relative .md file, or an https/mailto URL."},
			{Name: "icon", Type: TIcon, Doc: "Hand-drawn icon above the title (see the icon list)."},
			{Name: "cta", Type: TString, Doc: "Call-to-action text at the bottom of the card (e.g. \"Read the guide\"). Requires href."},
		},
		Example: "<Card title=\"Installation\" icon=\"rocket\" href=\"./installation.md\">\nPut the binary on your PATH.\n</Card>",
		Render:  renderCard,
	})

	columns := &Spec{
		Name:    "Columns",
		Summary: "Responsive grid. It becomes a single column on narrow screens.",
		Attrs: []AttrSpec{
			{Name: "cols", Type: TInt, Default: 2, Min: 1, Max: 4, Doc: "Number of columns (1 to 4)."},
		},
		Example: "<Columns cols={2}>\n<Card title=\"A\">...</Card>\n<Card title=\"B\">...</Card>\n</Columns>",
		Render:  renderColumns,
	}
	add(columns)
	add(&Spec{Name: "CardGroup", Summary: "Shorthand for <Columns> (the name Mintlify uses).", Attrs: columns.Attrs, AliasOf: "Columns", Example: "<CardGroup cols={2}>\n<Card title=\"A\">...</Card>\n</CardGroup>", Render: renderColumns})

	add(&Spec{
		Name:     "Tabs",
		Summary:  "Alternative content in tabs.",
		Children: ChildOnly,
		Allowed:  []string{"Tab"},
		Example:  "<Tabs>\n<Tab title=\"macOS\">\n...\n</Tab>\n<Tab title=\"Linux\">\n...\n</Tab>\n</Tabs>",
		Render:   renderTabs,
	})
	add(&Spec{
		Name:    "Tab",
		Summary: "One tab of <Tabs>.",
		Attrs:   []AttrSpec{{Name: "title", Type: TString, Required: true, Doc: "Tab label."}},
		Parent:  "Tabs",
		Example: "<Tab title=\"Linux\">\nContent in Markdown.\n</Tab>",
		Render:  renderTab,
	})

	add(&Spec{
		Name:     "CodeGroup",
		Summary:  "Several code blocks in tabs. Each label comes from title=\"...\" on the fence, or from the language.",
		Children: ChildOnly,
		Allowed:  []string{ChildCode},
		Example:  "<CodeGroup>\n```bash title=\"npm\"\nnpm i x\n```\n```bash title=\"pnpm\"\npnpm add x\n```\n</CodeGroup>",
		Render:   renderCodeGroup,
	})

	add(&Spec{
		Name:     "Steps",
		Summary:  "Numbered sequence of steps, with hand-drawn markers.",
		Children: ChildOnly,
		Allowed:  []string{"Step"},
		Example:  "<Steps>\n<Step title=\"Install\">\n...\n</Step>\n<Step title=\"Configure\">\n...\n</Step>\n</Steps>",
		Render:   renderSteps,
	})
	add(&Spec{
		Name:    "Step",
		Summary: "One step of <Steps>.",
		Attrs:   []AttrSpec{{Name: "title", Type: TString, Required: true, Doc: "Step title."}},
		Parent:  "Steps",
		Example: "<Step title=\"Install\">\nRun the installer.\n</Step>",
		Render:  renderStep,
	})

	add(&Spec{
		Name:    "Accordion",
		Summary: "Collapsible section.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Always-visible text."},
			{Name: "open", Type: TBool, Default: false, Doc: "Starts open."},
		},
		Example: "<Accordion title=\"Why Go?\">\nA single binary.\n</Accordion>",
		Render:  renderAccordion,
	})
	add(&Spec{
		Name:     "AccordionGroup",
		Summary:  "Groups several <Accordion> into one continuous block, like an FAQ.",
		Children: ChildOnly,
		Allowed:  []string{"Accordion"},
		Example:  "<AccordionGroup>\n<Accordion title=\"Question 1\">\nAnswer.\n</Accordion>\n<Accordion title=\"Question 2\">\nAnswer.\n</Accordion>\n</AccordionGroup>",
		Render:   renderAccordionGroup,
	})

	add(&Spec{
		Name:    "Button",
		Summary: "A link that looks like a button. The text comes from label or from the content. Consecutive buttons sit side by side.",
		Attrs: []AttrSpec{
			{Name: "href", Type: TURL, Required: true, Doc: "Destination."},
			{Name: "label", Type: TString, Doc: "Button text (or write the text between the tags)."},
			{Name: "variant", Type: TEnum, Enum: []string{"primary", "secondary"}, Default: "primary", Doc: "Style."},
		},
		Example:      "<Button href=\"/guide\" label=\"Get started\" />",
		Render:       renderButton,
		SkipChildren: true,
	})

	add(&Spec{
		Name:    "Hero",
		Summary: "Opening of an entry page, with a large title. Landing layout only. The content becomes the subtitle and can include <Button>.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Main title (becomes the page <h1>)."},
			{Name: "eyebrow", Type: TString, Doc: "Small text above the title."},
			{Name: "highlight", Type: TString, Doc: "Part of the title underlined with a hand-drawn stroke. Must appear exactly in title."},
			{Name: "image", Type: TAsset, Doc: "Image next to the text (SVG or PNG). With image, the hero has two columns."},
			{Name: "alt", Type: TString, Doc: "Image description. Required with image."},
			{Name: "align", Type: TEnum, Enum: []string{"center", "left"}, Default: "center", Doc: "Text alignment (without image)."},
		},
		Layouts: []string{"landing"},
		Example: "<Hero title=\"Docs that write themselves\" highlight=\"themselves\" eyebrow=\"Wavemark\">\nMarkdown becomes a site.\n\n<Button href=\"/guide\" label=\"Get started\" />\n</Hero>",
		Render:  renderHero,
	})

	add(&Spec{
		Name:    "Section",
		Summary: "Full-width band with its own background, to set apart part of a landing page (final call to action, testimonials, numbers).",
		Attrs: []AttrSpec{
			{Name: "tone", Type: TEnum, Enum: []string{"muted", "oat", "clay", "dark"}, Default: "muted", Doc: "Background color."},
			{Name: "align", Type: TEnum, Enum: []string{"left", "center"}, Default: "left", Doc: "Text alignment."},
		},
		Layouts: []string{"landing"},
		Example: "<Section tone=\"oat\" align=\"center\">\n## Ready to start?\n\n<Button href=\"/guide\" label=\"Read the guide\" />\n</Section>",
		Render:  renderSection,
	})

	add(&Spec{
		Name:    "Quote",
		Summary: "Pull quote with an author. For testimonials and epigraphs.",
		Attrs: []AttrSpec{
			{Name: "author", Type: TString, Doc: "Who said it."},
			{Name: "role", Type: TString, Doc: "Role, organization or source."},
		},
		Example: "<Quote author=\"Ada Lovelace\" role=\"Notes on the Analytical Engine\">\nThe Analytical Engine has no pretensions whatever to originate anything.\n</Quote>",
		Render:  renderQuote,
	})

	add(&Spec{
		Name:     "Stat",
		Summary:  "A prominent number with a label. Use several inside <Columns>.",
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "value", Type: TString, Required: true, Doc: "The number, as text (\"27\", \"180k\", \"99.9%\")."},
			{Name: "label", Type: TString, Required: true, Doc: "What the number measures."},
		},
		Example: "<Stat value=\"27\" label=\"essays published\" />",
		Render:  renderStat,
	})

	add(&Spec{
		Name:     "PageList",
		Summary:  "Automatic list of the pages in a folder, with date, author and description. For blogs and indexes.",
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "dir", Type: TDir, Default: "./", Doc: "Folder to list, relative to the page (default: the page's own folder)."},
			{Name: "sort", Type: TEnum, Enum: []string{"auto", "date", "title", "nav"}, Default: "auto", Doc: "Order: date (newest first), title, nav (navigation order) or auto (date when the pages have dates)."},
			{Name: "limit", Type: TInt, Default: 0, Min: 0, Max: 500, Doc: "Maximum number of items (0 = all)."},
			{Name: "style", Type: TEnum, Enum: []string{"list", "cards"}, Default: "list", Doc: "A list of rows or a grid of cards."},
		},
		Example: "<PageList dir=\"./blog\" sort=\"date\" limit={5} />",
		Render:  renderPageList,
	})

	add(&Spec{
		Name:     "Flow",
		Summary:  "A hand-drawn diagram of steps connected by arrows: pipelines, request flows, processes. Wraps vertically on narrow screens.",
		Children: ChildOnly,
		Allowed:  []string{"FlowStep"},
		Attrs: []AttrSpec{
			{Name: "caption", Type: TString, Doc: "Caption shown below the diagram."},
		},
		Example: "<Flow caption=\"From file to response\">\n<FlowStep title=\"Markdown\" icon=\"folder\">Your files</FlowStep>\n<FlowStep title=\"Parser\" icon=\"code\">Validated</FlowStep>\n<FlowStep title=\"HTTP\" icon=\"globe\">Served</FlowStep>\n</Flow>",
		Render:  renderFlow,
	})
	add(&Spec{
		Name:    "FlowStep",
		Summary: "One step of a <Flow>. The content is a short description.",
		Attrs: []AttrSpec{
			{Name: "title", Type: TString, Required: true, Doc: "Step name."},
			{Name: "icon", Type: TIcon, Doc: "Hand-drawn icon shown above the name."},
		},
		Parent:  "Flow",
		Example: "<FlowStep title=\"Parser\" icon=\"code\">goldmark + schemas</FlowStep>",
		Render:  renderFlowStep,
	})

	add(&Spec{
		Name:     "Figure",
		Summary:  "Image with a caption.",
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "src", Type: TAsset, Required: true, Doc: "Image file relative to the page, an absolute path in the site, or an https URL."},
			{Name: "alt", Type: TString, Required: true, Doc: "Image description for screen readers."},
			{Name: "caption", Type: TString, Doc: "Caption shown below."},
		},
		Example: "<Figure src=\"./architecture.png\" alt=\"Architecture diagram\" caption=\"How a request flows\" />",
		Render:  renderFigure,
	})

	add(&Spec{
		Name:         "Badge",
		Summary:      "Short label inside text, headings or table cells (\"New\", \"Beta\", \"v2.1\").",
		Inline:       true,
		Children:     ChildText,
		Attrs:        []AttrSpec{{Name: "color", Type: TEnum, Enum: badgeColors, Default: "gray", Doc: "Color."}},
		Example:      "<Badge color=\"clay\">New</Badge>",
		RenderInline: renderBadge,
	})
	add(&Spec{
		Name:     "Icon",
		Summary:  "Hand-drawn icon inside text.",
		Inline:   true,
		Children: ChildNone,
		Attrs: []AttrSpec{
			{Name: "name", Type: TIcon, Required: true, Doc: "Icon name."},
			{Name: "label", Type: TString, Doc: "Description for screen readers. Without label the icon is decorative."},
		},
		Example:      "<Icon name=\"sparkles\" />",
		RenderInline: renderIcon,
	})

	// Rules that need more than one attribute.
	registry["Card"].check = func(c *Core, diags *diag.List) {
		if c.Str("cta") != "" && c.Str("href") == "" {
			diags.Errorf(c.Line, c.Col, diag.CodeAttrValue, "Add href or remove cta.", "'cta' on <Card> requires 'href'")
		}
	}
	registry["Hero"].check = func(c *Core, diags *diag.List) {
		if h := c.Str("highlight"); h != "" && !strings.Contains(c.Str("title"), h) {
			diags.Errorf(c.Line, c.Col, diag.CodeAttrValue, "Copy the phrase exactly as it appears in title.", "highlight %q does not appear in the title of <Hero>", h)
		}
		if c.Str("image") != "" && c.Str("alt") == "" {
			diags.Errorf(c.Line, c.Col, diag.CodeAttrMissing, `Describe the image: alt="..."`, "<Hero image> requires the attribute 'alt'")
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

func renderFlow(w util.BufWriter, _ []byte, n *Node, entering bool, _ *RenderContext) {
	if entering {
		_, _ = w.WriteString(`<figure class="wm-flow"><ol class="wm-flow__steps">` + "\n")
		return
	}
	_, _ = w.WriteString("</ol>")
	if c := n.Str("caption"); c != "" {
		_, _ = w.WriteString(`<figcaption class="wm-flow__caption">` + esc(c) + "</figcaption>")
	}
	_, _ = w.WriteString("</figure>\n")
}

func renderFlowStep(w util.BufWriter, _ []byte, n *Node, entering bool, ctx *RenderContext) {
	if !entering {
		_, _ = w.WriteString("</div></li>\n")
		return
	}
	_, _ = w.WriteString(`<li class="wm-flow__step">`)
	if n.PreviousSibling() != nil {
		_, _ = w.WriteString(`<span class="wm-flow__arrow" aria-hidden="true">` + sketch.Arrow(sketch.Seed(n.Str("title")), sketch.Options{}) + "</span>")
	}
	_, _ = w.WriteString(`<div class="wm-flow__box">`)
	if icon := n.Str("icon"); icon != "" {
		_, _ = w.WriteString(`<span class="wm-flow__icon">` + sketch.Icon(icon, sketch.Options{Size: iconSize(ctx, 32, 40)}) + "</span>")
	}
	_, _ = w.WriteString(`<p class="wm-flow__title">` + esc(n.Str("title")) + `</p><div class="wm-flow__body">` + "\n")
}
