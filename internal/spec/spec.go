// Package spec generates the authoring reference printed by `wavemark spec`.
// It is built from the same schemas that validate content, so it cannot
// drift from the implementation.
package spec

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/config"
	"github.com/YutaViriato/Wavemark/internal/sketch"
)

// Markdown returns the full reference as Markdown.
func Markdown() string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }

	w("# Wavemark authoring reference\n\n")
	w("Wavemark turns a folder of Markdown files into a website. This document is the complete reference of the syntax it accepts. ")
	w("Always validate with `wavemark check <folder> --json` and fix until there are no errors.\n\n")

	w("## Folder structure\n\n")
	w("- Every `.md` file is a page. The route comes from the path: `guide/install.md` → `/guide/install`.\n")
	w("- `index.md` (or `README.md`) is the page of its folder: `guide/index.md` → `/guide`. The root `index.md` is the home page.\n")
	w("- Every folder is a section of the sidebar. A `_meta.yaml` in the folder sets its title, order and inherited settings.\n")
	w("- `wavemark.yaml` at the root configures the site.\n")
	w("- Other files (images, PDFs) are served at the same path: `img/a.png` → `/img/a.png`.\n")
	w("- Names starting with `.` and `node_modules` are ignored.\n\n")

	w("## Markdown\n\n")
	w("CommonMark with the GFM extensions: tables, ~~strikethrough~~, task lists, autolinks and footnotes (`[^1]`).\n\n")
	w("Wavemark rules:\n\n")
	w("- **Raw HTML is forbidden** (including `<br>`, `<div>`, `<img>`). Use Markdown or components. Comments `<!-- -->` are allowed and never rendered.\n")
	w("- **Fenced code blocks only** (```` ``` ````). A block indented by 4 spaces is an error.\n")
	w("- A fence takes a language and a title: ```` ```python title=\"main.py\" ````. More than 200 languages are highlighted (bash, python, typescript, go, rust, json, yaml, sql, html, css, diff…); `text` turns highlighting off.\n")
	w("- `mermaid` and math blocks are not rendered yet and show as code. For diagrams use `<Flow>`, or an SVG image with `<Figure>`.\n")
	w("- **One title per page**: `title` in the frontmatter or the first `# H1` of the body, never both. Use `##` and `###` for sections; they make up the table of contents.\n")
	w("- Anchors come from headings, lowercased and without accents: `## Advanced Setup` → `#advanced-setup`. To pin an anchor: `## Title {#my-anchor}`.\n")
	w("- Destinations accepted in links, `href` and `nav`: a relative `.md` file, an absolute route (`/guide`), an anchor (`#section`), `https://`, `mailto:` and `tel:`.\n")
	w("- **Internal links point to files**: `[Install](./install.md)`, `[Section](../guide/x.md#section)`. Absolute routes also work: `/guide/x`. Broken links and missing anchors are errors.\n")
	w("- Images: `![description](./img/a.png)`. The file must exist and the description must not be empty.\n")
	w("- Footnotes: `text[^1]` and, on its own line, `[^1]: the note`.\n\n")

	w("## Components\n\n")
	w("Components are PascalCase tags written like declarative JSX. There is no JavaScript.\n\n")
	w("Syntax rules:\n\n")
	w("1. Every tag sits **on its own line**: opening, closing and self-closing tags (`<Button ... />`).\n")
	w("2. An opening tag may span several lines between attributes, but not a blank line.\n")
	w("3. Content between tags is Markdown and may be indented; the indentation of the first content line is removed.\n")
	w("4. One-line form: `<Step title=\"Install\">Run the installer.</Step>`.\n")
	w("5. Values: quoted strings `title=\"...\"` (escape quotes as `\\\"`), JSON literals in braces `cols={3}` `open={false}`, or the bare name for true (`open`).\n")
	w("6. Unknown attributes, missing required ones and values of the wrong type are errors.\n")
	w("7. **Inline components** (`<Badge>`, `<Icon />`) are the exception: they go inside text, headings or table cells, and open and close on the same line: `## Export <Badge color=\"clay\">New</Badge>`.\n")
	w("8. `<Badge>` and `<Icon />` inside a heading are not part of its anchor: `## Export <Badge>New</Badge>` gives `#export`.\n\n")

	for _, s := range components.All() {
		w("### <%s>\n\n%s\n\n", s.Name, s.Summary)
		if s.Inline {
			w("Inline: use it inside a paragraph, heading or table cell.\n\n")
		}
		if len(s.Layouts) > 0 {
			w("Layouts: %s.\n\n", strings.Join(s.Layouts, ", "))
		}
		if s.Parent != "" {
			w("Must be directly inside `<%s>`.\n\n", s.Parent)
		}
		switch s.Children {
		case components.ChildNone:
			w("No content: use the self-closing form `<%s ... />`.\n\n", s.Name)
		case components.ChildOnly:
			var allowed []string
			for _, a := range s.Allowed {
				if a == components.ChildCode {
					allowed = append(allowed, "fenced code blocks")
				} else {
					allowed = append(allowed, "`<"+a+">`")
				}
			}
			w("Content: only %s.\n\n", strings.Join(allowed, " or "))
		case components.ChildText:
			w("Content: short text between the tags.\n\n")
		default:
			if !s.SkipChildren {
				w("Content: Markdown and other components.\n\n")
			}
		}
		if len(s.Attrs) > 0 {
			w("| Attribute | Type | Required | Default | Description |\n|---|---|---|---|---|\n")
			for _, a := range s.Attrs {
				typ := a.Type.String()
				if a.Type == components.TEnum {
					typ = strings.Join(quoteAll(a.Enum), " \\| ")
				}
				req := "no"
				if a.Required {
					req = "yes"
				}
				def := ""
				if a.Default != nil {
					def = "`" + jsonish(a.Default) + "`"
				}
				w("| `%s` | %s | %s | %s | %s |\n", a.Name, typ, req, def, a.Doc)
			}
			w("\n")
		}
		w("```mdx\n%s\n```\n\n", s.Example)
	}

	w("## Icons\n\n")
	w("Hand-drawn icons, used by `<Card icon>`, `<FlowStep icon>`, `<Icon name>` and the `icon` frontmatter key. Common Lucide, Font Awesome and Mintlify names also work (`rocket-launch`, `settings`, `magnifying-glass`, `envelope`…).\n\n")
	w("%s\n\n", "`"+strings.Join(sketch.Names(), "`, `")+"`")

	w("## Page frontmatter\n\n")
	w("A YAML block at the top of the file, between `---` lines. Every key is optional.\n\n")
	fieldTable(&b, config.Describe(config.Frontmatter{}))
	w("```yaml\n---\ntitle: Installation\ndescription: Install Wavemark in a minute.\norder: 1\n---\n```\n\n")
	w("A blog post or article:\n\n```yaml\n---\ntitle: Why we write slowly\ndescription: Notes on our research process.\nlayout: article\ndate: 2026-03-14\nauthor: Ada Lovelace\n---\n```\n\n")

	w("## _meta.yaml (per folder)\n\n")
	fieldTable(&b, config.Describe(config.Folder{}))
	w("`defaults` keys:\n\n")
	fieldTable(&b, config.Describe(config.Defaults{}))
	w("```yaml\ntitle: Guide\norder: [install, configuration, deploy]\ndefaults:\n  toc: true\n```\n\n")
	w("`defaults` also apply to the folder's own `index.md`. In the example below, the blog listing page sets `layout: docs` again in its own frontmatter.\n\n")
	w("A blog folder (newest posts first, all as articles):\n\n```yaml\ntitle: Blog\nsort: date\ndefaults:\n  layout: article\n```\n\n")

	w("## wavemark.yaml (root)\n\n")
	fieldTable(&b, config.Describe(config.Site{}))
	w("```yaml\ntitle: Acme Docs\ndescription: Documentation for Acme.\nlang: en\nlogo:\n  text: Acme\n  src: ./img/logo.svg\nnav:\n  - label: Guide\n    href: /guide\n  - label: API\n    href: ./api/index.md\nfooter:\n  text: © 2026 Acme\n  links:\n    - label: GitHub\n      href: https://github.com/acme\ndefaults:\n  layout: docs\n```\n\n")
	w("There are no theme, color, CSS or script keys: the look is the theme's.\n\n")

	w("## Configuration cascade\n\n")
	w("`wavemark.yaml › defaults` → each folder's `_meta.yaml › defaults`, outermost to innermost → frontmatter. ")
	w("The most specific level wins. Lists replace, they do not merge.\n\n")

	w("## Layouts\n\n")
	w("- `docs` (default): sidebar navigation, serif title, table of contents on the right, previous and next pages. For documentation.\n")
	w("- `article`: long-form reading without a sidebar, with chapters (`##`) on the left and date, author and section on the right. For essays, posts and news.\n")
	w("- `landing`: an entry page. Start with `<Hero>`; use `##` for sections, `<Columns>` with `<Card>` or `<Stat>` for grids and `<Section>` for full-width bands. The frontmatter `title` only becomes the browser tab title; the visible title is the `<Hero>` title.\n\n")
	w("## Motion\n\n")
	w("The theme animates on its own: the `<Hero>` title rises word by word, icons, `<Steps>` markers and `<Flow>` arrows draw themselves, blocks appear on scroll and page navigation has a transition. There is nothing to configure. Readers who ask their system for reduced motion see everything static.\n")
	return b.String()
}

func fieldTable(b *strings.Builder, fields []config.Field) {
	b.WriteString("| Key | Type | Description |\n|---|---|---|\n")
	for _, f := range fields {
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", f.Key, f.Type, f.Doc)
	}
	b.WriteString("\n")
}

func quoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = `"` + s + `"`
	}
	return out
}

func jsonish(v any) string {
	if s, ok := v.(string); ok {
		return `"` + s + `"`
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// JSON returns the component schemas as JSON, for tools.
func JSON() ([]byte, error) {
	type attr struct {
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Required bool     `json:"required"`
		Enum     []string `json:"enum,omitempty"`
		Default  any      `json:"default,omitempty"`
		Doc      string   `json:"doc"`
	}
	type comp struct {
		Name     string   `json:"name"`
		Summary  string   `json:"summary"`
		AliasOf  string   `json:"aliasOf,omitempty"`
		Parent   string   `json:"parent,omitempty"`
		Layouts  []string `json:"layouts,omitempty"`
		Children string   `json:"children"`
		Inline   bool     `json:"inline,omitempty"`
		Allowed  []string `json:"allowed,omitempty"`
		Attrs    []attr   `json:"attributes"`
		Example  string   `json:"example"`
	}
	var out []comp
	for _, s := range components.All() {
		c := comp{Name: s.Name, Summary: s.Summary, AliasOf: s.AliasOf, Parent: s.Parent, Layouts: s.Layouts, Inline: s.Inline, Example: s.Example, Attrs: []attr{}}
		switch s.Children {
		case components.ChildNone:
			c.Children = "none"
		case components.ChildOnly:
			c.Children = "only"
			c.Allowed = s.Allowed
		case components.ChildText:
			c.Children = "text"
		default:
			c.Children = "markdown"
		}
		for _, a := range s.Attrs {
			c.Attrs = append(c.Attrs, attr{Name: a.Name, Type: a.Type.String(), Required: a.Required, Enum: a.Enum, Default: a.Default, Doc: a.Doc})
		}
		out = append(out, c)
	}
	return json.MarshalIndent(map[string]any{"components": out, "icons": sketch.Names(), "layouts": config.Layouts}, "", "  ")
}
