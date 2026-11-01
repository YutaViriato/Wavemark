package markdown

import (
	stdhtml "html"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/util"
)

// htmlSanitizer renders raw HTML found in plain Markdown. Output is rebuilt
// from scratch: text is escaped, only allow-listed tags are emitted, only a
// few attributes survive (URLs are checked), and tags are kept balanced so a
// stray </div> cannot close the theme's containers. Anything else is shown
// as visible, escaped source.
type htmlSanitizer struct {
	url   func(string) string
	stack []string
}

var (
	htmlTokenRe = regexp.MustCompile(`(?s)<!--.*?-->|<(/?)([A-Za-z][A-Za-z0-9-]*)((?:\s+[^\s"'>/=]+(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'=<>` + "`" + `]+))?)*)\s*(/?)>`)
	htmlAttrRe  = regexp.MustCompile(`([^\s"'>/=]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>` + "`" + `]+)))?`)
)

// allowedTags are emitted without attributes, except those in allowedAttrs.
var allowedTags = map[string]bool{
	"a": true, "abbr": true, "b": true, "blockquote": true, "br": true, "code": true, "dd": true, "del": true,
	"details": true, "div": true, "dl": true, "dt": true, "em": true, "h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true, "hr": true, "i": true, "img": true, "ins": true, "kbd": true,
	"li": true, "mark": true, "ol": true, "p": true, "pre": true, "q": true, "s": true, "samp": true,
	"small": true, "span": true, "strong": true, "sub": true, "summary": true, "sup": true, "table": true,
	"tbody": true, "td": true, "tfoot": true, "th": true, "thead": true, "tr": true, "u": true, "ul": true,
	"var": true,
}

// tagAliases maps obsolete or wrapper tags to an allowed equivalent.
var tagAliases = map[string]string{"center": "div", "picture": "span", "tt": "code", "strike": "s"}

// droppedTags are removed silently (their text content stays).
var droppedTags = map[string]bool{"source": true, "font": true}

var voidTags = map[string]bool{"br": true, "hr": true, "img": true}

var alignTags = map[string]bool{"p": true, "div": true, "td": true, "th": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true}

func (h *htmlSanitizer) write(w util.BufWriter, raw string) {
	last := 0
	for _, m := range htmlTokenRe.FindAllStringSubmatchIndex(raw, -1) {
		h.text(w, raw[last:m[0]])
		last = m[1]
		if strings.HasPrefix(raw[m[0]:], "<!--") {
			continue
		}
		closing := m[3] > m[2]
		name := strings.ToLower(raw[m[4]:m[5]])
		attrs := raw[m[6]:m[7]]
		if a, ok := tagAliases[name]; ok {
			name = a
		}
		switch {
		case droppedTags[name]:
		case !allowedTags[name]:
			h.literal(w, raw[m[0]:m[1]])
		case closing:
			h.close(w, name)
		default:
			h.open(w, name, attrs)
		}
	}
	h.text(w, raw[last:])
}

func (h *htmlSanitizer) text(w util.BufWriter, s string) {
	if s == "" {
		return
	}
	// Normalise entities (&nbsp;, &copy;) and escape the rest.
	_, _ = w.WriteString(stdhtml.EscapeString(stdhtml.UnescapeString(s)))
}

// literal shows a tag that is not rendered as escaped source.
func (h *htmlSanitizer) literal(w util.BufWriter, tag string) {
	_, _ = w.WriteString(`<code class="wm-rawhtml">` + stdhtml.EscapeString(tag) + "</code>")
}

func (h *htmlSanitizer) open(w util.BufWriter, name, attrs string) {
	var b strings.Builder
	b.WriteString("<" + name)
	vals := map[string]string{}
	for _, am := range htmlAttrRe.FindAllStringSubmatch(attrs, -1) {
		k := strings.ToLower(am[1])
		vals[k] = stdhtml.UnescapeString(am[2] + am[3] + am[4])
		if am[2] == "" && am[3] == "" && am[4] == "" && !strings.Contains(am[0], "=") {
			vals[k] = "\x00bare"
		}
	}
	switch name {
	case "a":
		if href, ok := h.safeURL(vals["href"]); ok {
			b.WriteString(` href="` + stdhtml.EscapeString(href) + `"`)
			if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
				b.WriteString(` rel="noopener"`)
			}
		}
	case "img":
		src, ok := h.safeURL(vals["src"])
		if !ok {
			return // an image without a usable source shows nothing
		}
		b.WriteString(` src="` + stdhtml.EscapeString(src) + `"`)
		b.WriteString(` alt="` + stdhtml.EscapeString(vals["alt"]) + `"`)
		for _, k := range []string{"width", "height"} {
			if n, err := strconv.Atoi(strings.TrimSuffix(vals[k], "px")); err == nil && n > 0 && n < 10000 {
				b.WriteString(" " + k + `="` + strconv.Itoa(n) + `"`)
			}
		}
		b.WriteString(` loading="lazy" decoding="async"`)
	case "details":
		if _, ok := vals["open"]; ok {
			b.WriteString(" open")
		}
	}
	if t := vals["title"]; t != "" && t != "\x00bare" {
		b.WriteString(` title="` + stdhtml.EscapeString(t) + `"`)
	}
	if alignTags[name] {
		switch a := strings.ToLower(vals["align"]); a {
		case "left", "center", "right":
			b.WriteString(` align="` + a + `"`)
		}
	}
	b.WriteString(">")
	_, _ = w.WriteString(b.String())
	if !voidTags[name] {
		h.stack = append(h.stack, name)
	}
}

// close emits a closing tag only when the element is open, closing any
// elements opened inside it first.
func (h *htmlSanitizer) close(w util.BufWriter, name string) {
	for i := len(h.stack) - 1; i >= 0; i-- {
		if h.stack[i] != name {
			continue
		}
		for j := len(h.stack) - 1; j >= i; j-- {
			_, _ = w.WriteString("</" + h.stack[j] + ">")
		}
		h.stack = h.stack[:i]
		return
	}
}

func (h *htmlSanitizer) closeAll(w util.BufWriter) {
	for j := len(h.stack) - 1; j >= 0; j-- {
		_, _ = w.WriteString("</" + h.stack[j] + ">")
	}
	h.stack = nil
}

// safeURL maps a URL through the page's resolver and accepts relative URLs
// and the http, https, mailto and tel schemes.
func (h *htmlSanitizer) safeURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "\x00bare" {
		return "", false
	}
	for _, r := range raw {
		if r < 0x20 || r == 0x7f {
			return "", false
		}
	}
	if h.url != nil {
		raw = h.url(raw)
	}
	if i := strings.IndexAny(raw, ":/?#"); i >= 0 && raw[i] == ':' {
		switch strings.ToLower(raw[:i]) {
		case "http", "https", "mailto", "tel":
		default:
			return "", false
		}
	}
	return raw, true
}
