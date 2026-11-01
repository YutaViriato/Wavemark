package markdown

import (
	"bytes"
	"crypto/sha256"
	stdhtml "html"
	"strings"
	"sync"

	"github.com/YutaViriato/Wavemark/internal/components"
	"github.com/YutaViriato/Wavemark/internal/icons"
	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// RenderOptions configures the rendering of one page.
type RenderOptions struct {
	// URL maps a link or asset destination, as written, to its final href.
	URL func(dest string) string
	// T translates interface strings.
	T func(key string) string
}

// Render renders a document body to HTML.
func Render(doc *Document, opts RenderOptions) (string, error) {
	ctx := components.NewRenderContext(opts.URL, opts.T)
	nr := &nodeRenderer{ctx: ctx}
	r := renderer.NewRenderer(renderer.WithNodeRenderers(
		util.Prioritized(html.NewRenderer(), 1000),
		// The align attribute instead of style="", which the CSP blocks.
		util.Prioritized(extension.NewTableHTMLRenderer(extension.WithTableCellAlignMethod(extension.TableCellAlignAttribute)), 500),
		util.Prioritized(extension.NewStrikethroughHTMLRenderer(), 500),
		util.Prioritized(extension.NewTaskCheckBoxHTMLRenderer(), 500),
		util.Prioritized(extension.NewFootnoteHTMLRenderer(), 500),
		util.Prioritized(nr, 100),
	))
	var buf bytes.Buffer
	if err := r.Render(&buf, doc.Source, doc.Root); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type nodeRenderer struct {
	ctx *components.RenderContext
}

func (r *nodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(components.KindComponent, r.component)
	reg.Register(ast.KindFencedCodeBlock, r.code)
	reg.Register(ast.KindHeading, r.heading)
	reg.Register(ast.KindLink, r.link)
	reg.Register(ast.KindImage, r.image)
	reg.Register(ast.KindHTMLBlock, skip)
	reg.Register(ast.KindRawHTML, skip)
	reg.Register(ast.KindTextBlock, r.textBlock)
}

func skip(util.BufWriter, []byte, ast.Node, bool) (ast.WalkStatus, error) {
	return ast.WalkSkipChildren, nil
}

// textBlock renders goldmark text blocks (tight list items) and the empty
// blocks the component parser emits when it recovers from a syntax error.
func (r *nodeRenderer) textBlock(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering && n.NextSibling() != nil && n.FirstChild() != nil {
		_ = w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) component(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*components.Node)
	if n.Spec == nil || n.Spec.Render == nil {
		return ast.WalkContinue, nil
	}
	n.Spec.Render(w, src, n, entering, r.ctx)
	return ast.WalkContinue, nil
}

var formatter = chromahtml.New(
	chromahtml.WithClasses(true),
	chromahtml.ClassPrefix("hl-"),
	chromahtml.PreventSurroundingPre(true),
)

func (r *nodeRenderer) code(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.FencedCodeBlock)
	fi := components.Fence(n, src)
	var code bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		code.Write(seg.Value(src))
	}

	inGroup := false
	if p, ok := n.Parent().(*components.Node); ok && p.Name == "CodeGroup" {
		inGroup = true
	}
	label := fi.Label()
	if inGroup {
		id := r.ctx.Panels[n]
		_, _ = w.WriteString(`<div class="wm-tabs__panel" role="tabpanel" id="` + id.Panel + `" aria-labelledby="` + id.Tab + `">`)
		_, _ = w.WriteString(`<p class="wm-tabs__label">` + escape(label) + "</p>")
	} else {
		// Standalone blocks only get a header when the author names them;
		// otherwise the copy button floats over the code, as on the docs.
		_, _ = w.WriteString(`<div class="wm-code">`)
		if title := fi.Attrs["title"]; title != "" {
			_, _ = w.WriteString(`<div class="wm-code__header"><span class="wm-code__title">` + escape(title) + "</span>")
			components.WriteCopyButton(w, r.ctx)
			_, _ = w.WriteString("</div>")
		} else {
			components.WriteCopyButton(w, r.ctx)
		}
	}

	lang := strings.ToLower(fi.Lang)
	_, _ = w.WriteString(`<pre class="wm-code__pre" tabindex="0"><code`)
	if lang != "" {
		_, _ = w.WriteString(` class="language-` + escape(lang) + `"`)
	}
	_, _ = w.WriteString(">")
	highlight(w, lang, code.String())
	_, _ = w.WriteString("</code></pre></div>\n")
	return ast.WalkSkipChildren, nil
}

var langAlias = map[string]string{"mdx": "html", "curl": "bash", "sh": "bash", "shell": "bash", "zsh": "bash", "console": "bash", "py": "python", "js": "javascript", "ts": "typescript", "yml": "yaml", "md": "markdown", "rb": "ruby", "rs": "rust"}

// highlightCache memoizes highlighted HTML by language and code. Lexing is
// most of the render cost and code blocks rarely change between rebuilds.
var highlightCache = struct {
	sync.RWMutex
	m map[[32]byte]string
}{m: map[[32]byte]string{}}

const highlightCacheMax = 20000

func highlight(w util.BufWriter, lang, code string) {
	if a, ok := langAlias[lang]; ok {
		lang = a
	}
	key := sha256.Sum256([]byte(lang + "\x00" + code))
	highlightCache.RLock()
	out, ok := highlightCache.m[key]
	highlightCache.RUnlock()
	if !ok {
		out = highlightUncached(lang, code)
		highlightCache.Lock()
		if len(highlightCache.m) >= highlightCacheMax {
			clear(highlightCache.m)
		}
		highlightCache.m[key] = out
		highlightCache.Unlock()
	}
	_, _ = w.WriteString(out)
}

func highlightUncached(lang, code string) string {
	var lexer chroma.Lexer
	if lang != "" && lang != "text" && lang != "txt" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		return escape(code)
	}
	it, err := chroma.Coalesce(lexer).Tokenise(nil, code)
	if err != nil {
		return escape(code)
	}
	var b strings.Builder
	if err := formatter.Format(&b, styles.Fallback, it); err != nil {
		return escape(code)
	}
	return b.String()
}

func (r *nodeRenderer) heading(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	tag := "h" + string(rune('0'+n.Level))
	id := ""
	if v, ok := n.AttributeString("id"); ok {
		if b, ok := v.([]byte); ok {
			id = string(b)
		}
	}
	if entering {
		_, _ = w.WriteString("<" + tag)
		if id != "" {
			_, _ = w.WriteString(` id="` + escape(id) + `"`)
		}
		_, _ = w.WriteString(` class="wm-heading">`)
		return ast.WalkContinue, nil
	}
	if id != "" && n.Level > 1 {
		_, _ = w.WriteString(`<a class="wm-anchor" href="#` + escape(id) + `" aria-label="Link">` + icons.SVG("link") + "</a>")
	}
	_, _ = w.WriteString("</" + tag + ">\n")
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) link(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if !entering {
		_, _ = w.WriteString("</a>")
		return ast.WalkContinue, nil
	}
	href := r.ctx.URL(string(n.Destination))
	_, _ = w.WriteString(`<a href="`)
	if !html.IsDangerousURL([]byte(href)) {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape([]byte(href), false)))
	}
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="` + escape(string(util.UnescapePunctuations(n.Title))) + `"`)
	}
	if isExternal(href) {
		_, _ = w.WriteString(` rel="noopener"`)
	}
	_ = w.WriteByte('>')
	return ast.WalkContinue, nil
}

func (r *nodeRenderer) image(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	href := r.ctx.URL(string(n.Destination))
	_, _ = w.WriteString(`<img src="`)
	if !html.IsDangerousURL([]byte(href)) {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape([]byte(href), false)))
	}
	_, _ = w.WriteString(`" alt="` + escape(PlainText(n, src)) + `"`)
	if n.Title != nil {
		_, _ = w.WriteString(` title="` + escape(string(util.UnescapePunctuations(n.Title))) + `"`)
	}
	_, _ = w.WriteString(` loading="lazy" decoding="async">`)
	return ast.WalkSkipChildren, nil
}

func isExternal(href string) bool {
	return strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")
}

func escape(s string) string { return stdhtml.EscapeString(s) }
