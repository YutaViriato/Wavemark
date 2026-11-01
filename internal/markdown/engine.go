// Package markdown parses Wavemark Markdown (CommonMark + GFM + components)
// into a goldmark AST, analyses it (titles, headings, links, forbidden HTML)
// and renders it to HTML.
package markdown

import (
	"reflect"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Document is a parsed page body.
type Document struct {
	Source []byte
	Root   ast.Node
	state  *parseState
}

// Line returns the 1-based file line of a node, or 0.
func (d *Document) Line(n ast.Node) int {
	l, _ := d.Position(n)
	return l
}

// Position returns the 1-based file line and column of a node, or 0, 0.
func (d *Document) Position(n ast.Node) (int, int) {
	if c, ok := n.(interface{ Pos() int }); ok {
		if p := c.Pos(); p >= 0 {
			return d.state.position(p)
		}
	}
	if n.Type() == ast.TypeBlock && n.Lines().Len() > 0 {
		return d.state.position(n.Lines().At(0).Start)
	}
	for c := n.FirstChild(); c != nil; c = c.FirstChild() {
		if p, ok := c.(interface{ Pos() int }); ok && p.Pos() >= 0 {
			return d.state.position(p.Pos())
		}
	}
	return 0, 0
}

var md = newMarkdown()

func newMarkdown() goldmark.Markdown {
	codeBlock := reflect.TypeOf(parser.NewCodeBlockParser())
	var blocks []util.PrioritizedValue
	for _, v := range parser.DefaultBlockParsers() {
		// Indented code blocks are disabled: content inside components is
		// often indented, and only fenced code is allowed.
		if reflect.TypeOf(v.Value) == codeBlock {
			continue
		}
		blocks = append(blocks, v)
	}
	// Before the HTML block parser (900) so capitalised tags are components.
	blocks = append(blocks, util.Prioritized(&componentParser{}, 850))

	p := parser.NewParser(
		parser.WithBlockParsers(blocks...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
		parser.WithAutoHeadingID(),
		parser.WithAttribute(),
	)
	return goldmark.New(
		goldmark.WithParser(p),
		goldmark.WithExtensions(extension.GFM, extension.Footnote),
	)
}

// Parse parses a page body. lineOffset is the number of file lines before
// the body (the frontmatter), so diagnostics point at the right line.
func Parse(body []byte, lineOffset int, diags *diag.List) *Document {
	st := newParseState(body, lineOffset, diags)
	pc := parser.NewContext(parser.WithIDs(newIDs()))
	pc.Set(stateKey, st)
	root := md.Parser().Parse(text.NewReader(body), parser.WithContext(pc))
	return &Document{Source: body, Root: root, state: st}
}
