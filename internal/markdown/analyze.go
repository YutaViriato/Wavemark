package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// Heading is a section heading with its anchor.
type Heading struct {
	Level int
	ID    string
	Text  string
}

// RefType says how a reference must be resolved.
type RefType int

const (
	RefLink  RefType = iota // must resolve to a page, anchor or external URL
	RefAsset                // must resolve to a file or external URL
	RefDir                  // must resolve to a folder of the site
)

// Ref is a link or asset reference found in a page.
type Ref struct {
	Type      RefType
	Dest      string
	Line, Col int
	Where     string // human description, e.g. "link", "imagem", "<Card href>"
}

// Analysis is what the site needs to know about a page body.
type Analysis struct {
	Title     string // from the first H1 (removed from the body), if any
	HeroTitle string
	Headings  []Heading // h2–h6 in order
	IDs       map[string]bool
	Refs      []Ref
	Plain     string // plain-text excerpt for descriptions
}

var voidTagRe = regexp.MustCompile(`(?i)^<\s*(br|hr|wbr)\b`)

// footnoteRefRe finds footnote references left as text, which means the
// footnote has no definition.
var footnoteRefRe = regexp.MustCompile(`\[\^[^\]\s]+\]`)

// Analyze validates a parsed body and extracts headings, title and
// references. When the frontmatter has a title, an H1 in the body is an error.
func Analyze(doc *Document, layout string, frontmatterTitle bool, diags *diag.List) *Analysis {
	a := &Analysis{IDs: map[string]bool{}}
	var h1s []*ast.Heading
	src := doc.Source

	_ = ast.Walk(doc.Root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := node.(type) {
		case *components.Node:
			components.Validate(n, layout, doc.Line, diags)
			components.Check(&n.Core, diags)
			checkComponent(doc, n, diags)
			a.addRefs(&n.Core)
			if n.Name == "Hero" && a.HeroTitle == "" {
				a.HeroTitle = n.Str("title")
			}
		case *components.InlineNode:
			components.ValidateInline(n, diags)
			components.Check(&n.Core, diags)
			a.addRefs(&n.Core)
		case *ast.Heading:
			if n.Level == 1 {
				h1s = append(h1s, n)
				break
			}
			id, _ := n.AttributeString("id")
			ids, _ := id.([]byte)
			h := Heading{Level: n.Level, ID: string(ids), Text: PlainText(n, src)}
			a.Headings = append(a.Headings, h)
			a.IDs[h.ID] = true
		case *ast.Link:
			l, c := doc.Position(n)
			a.Refs = append(a.Refs, Ref{Type: RefLink, Dest: string(n.Destination), Line: l, Col: c, Where: "link"})
		case *ast.Image:
			l, c := doc.Position(n)
			a.Refs = append(a.Refs, Ref{Type: RefAsset, Dest: string(n.Destination), Line: l, Col: c, Where: "image"})
			if strings.TrimSpace(PlainText(n, src)) == "" {
				diags.Warnf(l, c, diag.CodeAltMissing, "Describe the image between the brackets: ![flow diagram](./flow.svg).", "image without alt text")
			}
		case *ast.HTMLBlock:
			var raw []byte
			if n.Lines().Len() > 0 {
				seg := n.Lines().At(0)
				raw = seg.Value(src)
			}
			reportHTML(doc, n, raw, true, diags)
		case *ast.RawHTML:
			var raw []byte
			if n.Segments.Len() > 0 {
				seg := n.Segments.At(0)
				raw = seg.Value(src)
			}
			reportHTML(doc, n, raw, false, diags)
		case *ast.FencedCodeBlock:
			checkFence(doc, n, diags)
		case *ast.Paragraph:
			if m := footnoteRefRe.FindString(PlainText(n, src)); m != "" {
				l, c := doc.Position(n)
				diags.Warnf(l, c, diag.CodeLinkBroken, fmt.Sprintf("Define the footnote on its own line: %s: footnote text.", m), "footnote %s has no definition", m)
			}
		case *ast.CodeBlock:
			l, _ := doc.Position(n)
			diags.Errorf(l, 0, diag.CodeIndentedCode, "Use a fenced block (```language ... ```). If it was meant as text, remove the 4-space indentation.", "code block indented with 4 spaces")
		}
		return ast.WalkContinue, nil
	})

	for i, h := range h1s {
		l, c := doc.Position(h)
		switch {
		case i == 0 && frontmatterTitle:
			diags.Errorf(l, c, diag.CodeTitleDuplicate, "Remove the '# heading' from the body or the 'title' key from the frontmatter.", "the title is in the frontmatter and also in an H1 in the body")
		case i == 0:
			a.Title = PlainText(h, src)
			if id, ok := h.AttributeString("id"); ok {
				if b, ok := id.([]byte); ok {
					a.IDs[string(b)] = true
				}
			}
		default:
			diags.Errorf(l, c, diag.CodeTitleMultipleH1, "Use '##' for sections; a page has one title.", "more than one H1 on the page")
		}
	}
	// The first H1 becomes the page title rendered by the layout.
	if len(h1s) > 0 && !frontmatterTitle && h1s[0].Parent() == doc.Root {
		doc.Root.RemoveChild(doc.Root, h1s[0])
		if layout == "landing" {
			l, c := doc.Position(h1s[0])
			diags.Warnf(l, c, diag.CodeTitleHidden, "Use <Hero title=\"...\"> for the visible title of a landing page.", "the landing layout does not show the H1; it only becomes the page <title>")
		}
	}
	a.Plain = excerpt(doc.Root, src)
	return a
}

func (a *Analysis) addRefs(c *components.Core) {
	for _, la := range components.LinkAttrs(c) {
		t := RefLink
		switch la.Type {
		case components.TAsset:
			t = RefAsset
		case components.TDir:
			t = RefDir
		}
		a.Refs = append(a.Refs, Ref{Type: t, Dest: la.Attr.Value.(string), Line: la.Attr.Line, Col: la.Attr.Col, Where: fmt.Sprintf("<%s %s>", c.Name, la.Attr.Name)})
	}
}

func reportHTML(doc *Document, n ast.Node, raw []byte, block bool, diags *diag.List) {
	s := strings.TrimSpace(string(raw))
	if strings.HasPrefix(s, "<!--") {
		return // comments are allowed and never rendered
	}
	if !block && strings.HasPrefix(s, "</") {
		return // the closing half of inline HTML is reported with its opening tag
	}
	l, c := doc.Position(n)
	hint := "Raw HTML is not allowed. Use Markdown or a component (run 'wavemark spec')."
	switch {
	case voidTagRe.MatchString(s):
		hint = "For a line break, end the line with '\\'. For a new paragraph, leave a blank line."
	case looksLikeComponent([]byte(s)):
		name := strings.TrimLeft(s, "</")
		if i := strings.IndexAny(name, " \t/>"); i >= 0 {
			name = name[:i]
		}
		if spec := components.Lookup(name); spec != nil && spec.Inline {
			hint = fmt.Sprintf("Inline components open and close on the same line: %s", spec.Example)
		} else {
			hint = "Block components start on their own line, with no text before the tag."
		}
	case strings.HasPrefix(strings.ToLower(s), "<img"):
		hint = `Use ![alt](src) or <Figure src="..." alt="..." />.`
	case strings.HasPrefix(strings.ToLower(s), "<details"):
		hint = `Use <Accordion title="...">.`
	case strings.HasPrefix(strings.ToLower(s), "<a "):
		hint = "Use [text](destination)."
	case strings.HasPrefix(strings.ToLower(s), "<div"):
		hint = "Use <Columns>, <Card> or <Callout> for structure."
	}
	snippet := s
	if i := strings.IndexByte(snippet, '\n'); i >= 0 {
		snippet = snippet[:i]
	}
	if len(snippet) > 40 {
		snippet = snippet[:40] + "…"
	}
	diags.Errorf(l, c, diag.CodeHTMLForbidden, hint, "raw HTML is not allowed: %s", snippet)
}

func checkFence(doc *Document, n *ast.FencedCodeBlock, diags *diag.List) {
	fi := components.Fence(n, doc.Source)
	l, _ := doc.Position(n)
	switch strings.ToLower(fi.Lang) {
	case "mermaid", "math", "latex", "katex", "plantuml", "dot", "graphviz":
		diags.Warnf(l, 0, diag.CodeFenceLang, "Diagrams and math are not rendered yet; the block shows as code. For diagrams use <Flow>, or an SVG image with <Figure>.", "'%s' blocks render as code, not as a diagram", fi.Lang)
	}
	for _, bad := range fi.Bad {
		diags.Errorf(l, 0, diag.CodeComponentSyntax, "Use the form ```language title=\"name\".", "invalid attributes on the code fence: %s", bad)
	}
	for k := range fi.Attrs {
		if !slices.Contains(components.FenceAttrs, k) {
			diags.Errorf(l, 0, diag.CodeAttrUnknown, "Attributes accepted on the fence: "+strings.Join(components.FenceAttrs, ", ")+".", "unknown attribute '%s' on the code fence", k)
		}
	}
}

// checkComponent applies rules that need the surrounding tree.
func checkComponent(doc *Document, n *components.Node, diags *diag.List) {
	if n.Spec == nil {
		return
	}
	if n.Name == "Button" {
		text := components.TextOf(n, doc.Source)
		switch {
		case n.Str("label") == "" && text == "":
			diags.Errorf(n.Line, n.Col, diag.CodeAttrMissing, "Example: <Button href=\"/guide\" label=\"Get started\" /> or <Button href=\"/guide\">Get started</Button>", "<Button> needs text: use label or write the text between the tags")
		case n.Str("label") != "" && text != "":
			diags.Errorf(n.Line, n.Col, diag.CodeAttrValue, "Use label or the content, not both.", "<Button> has both label and content")
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if _, ok := c.(*ast.Paragraph); !ok || c.NextSibling() != nil {
				diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, "A button's text is a single plain line.", "<Button> accepts a single line of text")
				break
			}
		}
		return
	}
	if n.Name != "Card" {
		return
	}
	href := n.Str("href")
	if href == "" {
		return
	}
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || c == ast.Node(n) {
			return ast.WalkContinue, nil
		}
		switch c.(type) {
		case *ast.Link, *components.Node:
			if cc, ok := c.(*components.Node); ok && cc.Name != "Button" {
				return ast.WalkContinue, nil
			}
			l, col := doc.Position(c)
			diags.Errorf(l, col, diag.CodeComponentChild, "The whole card is already a link. Remove the inner link or the href of <Card>.", "link inside <Card href>")
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
}

// PlainText returns the text content of a node, without markup.
func PlainText(n ast.Node, src []byte) string {
	var b bytes.Buffer
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Value(src))
			if t.SoftLineBreak() || t.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(t.Value)
		case *ast.CodeSpan:
			for cc := t.FirstChild(); cc != nil; cc = cc.NextSibling() {
				if tt, ok := cc.(*ast.Text); ok {
					b.Write(tt.Value(src))
				}
			}
			return ast.WalkSkipChildren, nil
		case *components.InlineNode:
			return ast.WalkSkipChildren, nil // badges and icons are decoration, not part of the title
		case *ast.RawHTML, *east.FootnoteLink:
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(strings.Fields(b.String()), " ")
}

// excerpt returns the text of the first paragraph, for llms.txt and meta
// descriptions when the page has none.
func excerpt(root ast.Node, src []byte) string {
	for c := root.FirstChild(); c != nil; c = c.NextSibling() {
		if p, ok := c.(*ast.Paragraph); ok {
			s := PlainText(p, src)
			if r := []rune(s); len(r) > 200 {
				s = strings.TrimSpace(string(r[:200])) + "…"
			}
			return s
		}
	}
	return ""
}
