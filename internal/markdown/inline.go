package markdown

import (
	"bytes"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// inlineParser parses inline components inside paragraphs, headings and
// table cells: <Badge color="clay">Novo</Badge> and <Icon name="rocket" />.
// Anything it cannot parse falls through to the raw HTML parser, which
// reports it.
type inlineParser struct{}

func (p *inlineParser) Trigger() []byte { return []byte{'<'} }

func (p *inlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	st := getState(pc)
	line, seg := block.PeekLine()
	if st == nil || len(line) < 2 || line[0] != '<' || !isUpper(line[1]) {
		return nil
	}
	src := block.Source()
	t, terr := lexTag(src, seg.Start)
	if terr != nil || t.Closing {
		return nil
	}
	spec := components.Lookup(t.Name)
	if spec == nil || !spec.Inline {
		return nil
	}
	l, c := st.position(seg.Start)
	n := components.NewInline(t.Name, l, c)
	for _, a := range t.Attrs {
		al, ac := st.position(a.Off)
		n.Attrs = append(n.Attrs, components.Attr{Name: a.Name, Value: a.Value, Form: a.Form, Line: al, Col: ac})
	}
	if t.SelfClosing {
		block.Advance(t.End - seg.Start)
		return n
	}
	rest := line[t.End-seg.Start:]
	closeTag := []byte("</" + t.Name + ">")
	i := bytes.Index(rest, closeTag)
	if i < 0 {
		return nil // reported as raw HTML with a hint about closing on the same line
	}
	n.Content = string(rest[:i])
	block.Advance(t.End - seg.Start + i + len(closeTag))
	return n
}
