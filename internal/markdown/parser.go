package markdown

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// parseState is shared by the parsers of one document through the goldmark
// parser context.
type parseState struct {
	diags      *diag.List
	lineStarts []int // byte offset of each line start in the body
	lineOffset int   // lines before the body in the page file (frontmatter)
	src        []byte
	open       map[*components.Node]*openInfo
}

type openInfo struct {
	contentIndent int
	indentSet     bool
}

var stateKey = parser.NewContextKey()

func newParseState(src []byte, lineOffset int, diags *diag.List) *parseState {
	starts := []int{0}
	for i, c := range src {
		if c == '\n' {
			starts = append(starts, i+1)
		}
	}
	return &parseState{diags: diags, lineStarts: starts, lineOffset: lineOffset, src: src, open: map[*components.Node]*openInfo{}}
}

// position converts a body byte offset to a 1-based file line and column.
func (s *parseState) position(off int) (line, col int) {
	if off < 0 {
		return 0, 0
	}
	lo, hi := 0, len(s.lineStarts)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if s.lineStarts[mid] <= off {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	start := s.lineStarts[lo]
	end := min(off, len(s.src))
	return lo + 1 + s.lineOffset, utf8.RuneCount(s.src[start:end]) + 1
}

func getState(pc parser.Context) *parseState {
	s, _ := pc.Get(stateKey).(*parseState)
	return s
}

// componentParser parses block-level component tags.
//
// Rules (see CLAUDE.md):
//   - tags sit on their own line; an opening tag may span several lines;
//   - `<Name ...>text</Name>` on one line is the inline-content form;
//   - content between tags is Markdown and may be indented; the component
//     strips the indentation of its first content line from every line;
//   - a closing tag belongs to the innermost open component, unless a fenced
//     code block is open inside it (then it is code), or it names an outer
//     component that nothing deeper matches (then the deeper ones are
//     reported as unclosed).
type componentParser struct{}

func (p *componentParser) Trigger() []byte { return []byte{'<'} }

func (p *componentParser) CanInterruptParagraph() bool { return true }

func (p *componentParser) CanAcceptIndentedLine() bool { return false }

func (p *componentParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	st := getState(pc)
	line, seg := reader.PeekLine()
	pos := pc.BlockOffset()
	if st == nil || pos < 0 || pos >= len(line) || !looksLikeComponent(line[pos:]) {
		return nil, parser.NoChildren
	}
	abs := seg.Start + pos
	src := reader.Source()

	t, terr := lexTag(src, abs)
	if terr != nil {
		l, c := st.position(terr.Off)
		st.diags.Errorf(l, c, diag.CodeComponentSyntax, terr.Hint, "%s", terr.Msg)
		reader.AdvanceToEOL()
		return ast.NewTextBlock(), parser.NoChildren
	}
	l, c := st.position(abs)

	if t.Closing {
		st.diags.Errorf(l, c, diag.CodeComponentStray, fmt.Sprintf("Remova </%s> ou adicione <%s> antes dele.", t.Name, t.Name), "</%s> sem um <%s> aberto", t.Name, t.Name)
		reader.AdvanceToEOL()
		return ast.NewTextBlock(), parser.NoChildren
	}

	n := components.NewNode(t.Name, l, c)
	for _, a := range t.Attrs {
		al, ac := st.position(a.Off)
		n.Attrs = append(n.Attrs, components.Attr{Name: a.Name, Value: a.Value, Form: a.Form, Line: al, Col: ac})
	}

	advanceTo(reader, t.End)
	rest, restSeg := reader.PeekLine()
	trimmed := bytes.TrimSpace(rest)

	if t.SelfClosing {
		n.SelfClosing, n.Closed = true, true
		if len(trimmed) > 0 {
			rl, rc := st.position(restSeg.Start)
			st.diags.Errorf(rl, rc, diag.CodeComponentSyntax, "Coloque o texto em uma linha própria, depois da tag.", "texto depois de <%s />", t.Name)
		}
		reader.AdvanceToEOL()
		return n, parser.NoChildren
	}

	if len(trimmed) == 0 {
		st.open[n] = &openInfo{}
		return n, parser.HasChildren
	}

	// Inline-content form: <Name ...>text</Name>
	closeTag := []byte("</" + t.Name + ">")
	if bytes.HasSuffix(trimmed, closeTag) {
		lead := len(rest) - len(bytes.TrimLeft(rest, " \t"))
		start := restSeg.Start + lead
		stop := start + len(trimmed) - len(closeTag)
		inner := bytes.TrimSpace(src[start:stop])
		if len(inner) > 0 {
			para := ast.NewParagraph()
			para.Lines().Append(text.NewSegment(start, stop))
			n.AppendChild(n, para)
		}
		n.Closed = true
		reader.AdvanceToEOL()
		return n, parser.NoChildren
	}

	rl, rc := st.position(restSeg.Start)
	st.diags.Errorf(rl, rc, diag.CodeComponentSyntax,
		fmt.Sprintf("Coloque o conteúdo nas linhas seguintes e feche com </%s> em linha própria, ou use a forma de uma linha <%s ...>texto</%s>.", t.Name, t.Name, t.Name),
		"conteúdo na mesma linha da tag <%s> sem o fechamento </%s>", t.Name, t.Name)
	reader.AdvanceToEOL()
	st.open[n] = &openInfo{}
	return n, parser.HasChildren
}

// advanceTo moves the reader to the absolute offset off, possibly across
// lines (multi-line opening tags).
func advanceTo(reader text.Reader, off int) {
	for {
		line, seg := reader.PeekLine()
		if line == nil {
			return
		}
		if off < seg.Stop || seg.Stop >= len(reader.Source()) {
			if d := off - seg.Start; d > 0 {
				reader.Advance(d)
			}
			return
		}
		reader.AdvanceLine()
	}
}

func (p *componentParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	n, ok := node.(*components.Node)
	if !ok || n.Closed {
		return parser.Close // recovery placeholder, self-closing or one-line form
	}
	st := getState(pc)
	line, _ := reader.PeekLine()
	if line == nil {
		return parser.Close
	}
	if util.IsBlank(line) {
		return parser.Continue | parser.HasChildren
	}
	trimmed := bytes.TrimSpace(line)
	if name, ok := closingName(trimmed); ok && p.owns(n, name, pc) {
		if name != n.Name {
			_, seg := reader.PeekLine()
			lead := len(line) - len(bytes.TrimLeft(line, " \t"))
			l, c := st.position(seg.Start + lead)
			st.diags.Errorf(l, c, diag.CodeComponentMismatch, fmt.Sprintf("Feche <%s> (aberto na linha %d) com </%s>.", n.Name, n.Line, n.Name), "esperava </%s>, encontrou </%s>", n.Name, name)
		}
		n.Closed = true
		reader.AdvanceToEOL()
		return parser.Close
	}

	info := st.open[n]
	if info == nil {
		info = &openInfo{}
		st.open[n] = info
	}
	w, _ := util.IndentWidth(line, reader.LineOffset())
	if !info.indentSet {
		info.contentIndent, info.indentSet = w, true
	}
	if strip := min(w, info.contentIndent); strip > 0 {
		pos, padding := util.IndentPosition(line, reader.LineOffset(), strip)
		if pos >= 0 {
			reader.AdvanceAndSetPadding(pos, padding)
		}
	}
	return parser.Continue | parser.HasChildren
}

// owns decides whether n should consume a closing tag for name.
func (p *componentParser) owns(n *components.Node, name string, pc parser.Context) bool {
	blocks := pc.OpenedBlocks()
	idx := -1
	for i, b := range blocks {
		if b.Node == ast.Node(n) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return true
	}
	deeperComponent := false
	deeperSameName := false
	for _, b := range blocks[idx+1:] {
		switch d := b.Node.(type) {
		case *ast.FencedCodeBlock:
			return false // the closing tag is code content
		case *components.Node:
			if d.Closed {
				continue
			}
			deeperComponent = true
			if d.Name == name {
				deeperSameName = true
			}
		}
	}
	if !deeperComponent {
		return true
	}
	return name == n.Name && !deeperSameName
}

func closingName(b []byte) (string, bool) {
	if len(b) < 4 || b[0] != '<' || b[1] != '/' || b[len(b)-1] != '>' || !isUpper(b[2]) {
		return "", false
	}
	name := bytes.TrimSpace(b[2 : len(b)-1])
	for _, c := range name {
		if !isNameChar(c) {
			return "", false
		}
	}
	return string(name), true
}

func (p *componentParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	n, ok := node.(*components.Node)
	if !ok {
		return
	}
	st := getState(pc)
	if st != nil {
		delete(st.open, n)
	}
	if !n.Closed && st != nil {
		st.diags.Errorf(n.Line, n.Col, diag.CodeComponentUnclosed, fmt.Sprintf("Adicione </%s> em uma linha própria depois do conteúdo, ou use <%s ... /> se ele não tiver conteúdo.", n.Name, n.Name), "<%s> aberto e nunca fechado", n.Name)
		n.Closed = true
	}
}
