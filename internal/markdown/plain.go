package markdown

import (
	"bytes"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// plainMD parses ordinary GitHub-flavoured Markdown, the kind agents and
// people write in any project: no Wavemark components and no {#id}
// attribute syntax, so every file renders the way it reads on GitHub.
var plainMD = goldmark.New(
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithExtensions(extension.GFM, extension.Footnote),
)

// ParsePlain parses a body as plain GitHub-flavoured Markdown. It is the
// lenient counterpart of Parse, used to browse Markdown that was not written
// for Wavemark. Render it with RenderOptions.Plain set.
func ParsePlain(body []byte, lineOffset int) *Document {
	st := newParseState(body, lineOffset, &diag.List{})
	pc := parser.NewContext(parser.WithIDs(newIDs()))
	root := plainMD.Parser().Parse(text.NewReader(body), parser.WithContext(pc))
	return &Document{Source: body, Root: root, state: st}
}

// plainRenderer renders the nodes that strict pages forbid: indented code,
// raw HTML (sanitised) and GitHub alerts.
type plainRenderer struct {
	nr   *nodeRenderer
	html *htmlSanitizer
}

func (r *plainRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindCodeBlock, r.codeBlock)
	reg.Register(ast.KindHTMLBlock, r.htmlBlock)
	reg.Register(ast.KindRawHTML, r.rawHTML)
	reg.Register(ast.KindDocument, r.document)
}

func (r *plainRenderer) document(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		r.html.closeAll(w)
	}
	return ast.WalkContinue, nil
}

func (r *plainRenderer) codeBlock(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	var code bytes.Buffer
	for i := 0; i < node.Lines().Len(); i++ {
		seg := node.Lines().At(i)
		code.Write(seg.Value(src))
	}
	_, _ = w.WriteString(`<div class="wm-code">`)
	components.WriteCopyButton(w, r.nr.ctx)
	_, _ = w.WriteString(`<pre class="wm-code__pre" tabindex="0"><code>`)
	_, _ = w.WriteString(escape(code.String()))
	_, _ = w.WriteString("</code></pre></div>\n")
	return ast.WalkSkipChildren, nil
}

func (r *plainRenderer) htmlBlock(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.HTMLBlock)
	var raw bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		raw.Write(seg.Value(src))
	}
	if n.HasClosure() {
		raw.Write(n.ClosureLine.Value(src))
	}
	r.html.write(w, raw.String())
	_ = w.WriteByte('\n')
	return ast.WalkSkipChildren, nil
}

func (r *plainRenderer) rawHTML(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.RawHTML)
	var raw strings.Builder
	for i := 0; i < n.Segments.Len(); i++ {
		seg := n.Segments.At(i)
		raw.Write(seg.Value(src))
	}
	r.html.write(w, raw.String())
	return ast.WalkSkipChildren, nil
}

// alertKinds maps GitHub alert markers to callout types and titles.
var alertKinds = map[string][2]string{
	"[!NOTE]":      {"note", "Note"},
	"[!TIP]":       {"tip", "Tip"},
	"[!IMPORTANT]": {"info", "Important"},
	"[!WARNING]":   {"warning", "Warning"},
	"[!CAUTION]":   {"danger", "Caution"},
}

// convertAlerts turns GitHub alerts ("> [!NOTE]" blockquotes) into callouts.
func convertAlerts(root ast.Node, src []byte) {
	var quotes []*ast.Blockquote
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if q, ok := n.(*ast.Blockquote); ok && entering {
			quotes = append(quotes, q)
		}
		return ast.WalkContinue, nil
	})
	for _, q := range quotes {
		p, ok := q.FirstChild().(*ast.Paragraph)
		if !ok || p.Lines().Len() == 0 {
			continue
		}
		first := p.Lines().At(0)
		kind, ok := alertKinds[strings.ToUpper(strings.TrimSpace(string(first.Value(src))))]
		if !ok {
			continue
		}
		// Drop the marker: the text nodes of the paragraph's first line.
		for c := p.FirstChild(); c != nil; {
			t, ok := c.(*ast.Text)
			if !ok || t.Segment.Start >= first.Stop {
				break
			}
			next := c.NextSibling()
			p.RemoveChild(p, c)
			c = next
		}
		if p.FirstChild() == nil {
			q.RemoveChild(q, p)
		}
		callout := components.NewNode("Callout", 0, 0)
		callout.Spec = components.Lookup("Callout")
		callout.Values = map[string]any{"type": kind[0], "title": kind[1]}
		callout.Closed = true
		for c := q.FirstChild(); c != nil; {
			next := c.NextSibling()
			q.RemoveChild(q, c)
			callout.AppendChild(callout, c)
			c = next
		}
		q.Parent().ReplaceChild(q.Parent(), q, callout)
	}
}
