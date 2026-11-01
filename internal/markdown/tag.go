package markdown

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/components"
)

// tag is a lexed component tag.
type tag struct {
	Name        string
	Closing     bool // </Name>
	SelfClosing bool // <Name ... />
	Attrs       []lexedAttr
	End         int // offset just past '>'
}

type lexedAttr struct {
	Name  string
	Value any
	Form  components.AttrForm
	Off   int // offset of the attribute name
}

type tagError struct {
	Off  int
	Msg  string
	Hint string
}

func (e *tagError) Error() string { return e.Msg }

// maxTagBytes bounds how far a multi-line opening tag may extend.
const maxTagBytes = 8 << 10

func isUpper(c byte) bool { return c >= 'A' && c <= 'Z' }
func isNameChar(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
}
func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

// looksLikeComponent reports whether b starts a component tag: '<' or '</'
// followed by an uppercase letter.
func looksLikeComponent(b []byte) bool {
	if len(b) < 2 || b[0] != '<' {
		return false
	}
	if b[1] == '/' {
		return len(b) > 2 && isUpper(b[2])
	}
	return isUpper(b[1])
}

// lexTag parses a component tag starting at src[start] == '<'. Offsets in
// the result and in errors are absolute offsets into src. A tag may span
// several lines but not a blank line.
func lexTag(src []byte, start int) (*tag, *tagError) {
	i := start + 1
	t := &tag{}
	if i < len(src) && src[i] == '/' {
		t.Closing = true
		i++
	}
	nameStart := i
	for i < len(src) && (isNameChar(src[i]) || src[i] == '.') {
		i++
	}
	t.Name = string(src[nameStart:i])
	if strings.ContainsAny(t.Name, ".-") {
		return nil, &tagError{Off: nameStart, Msg: fmt.Sprintf("invalid component name <%s>", t.Name), Hint: "Component names are single PascalCase words, like <Card>."}
	}
	limit := min(len(src), start+maxTagBytes)
	for {
		// Skip whitespace. A tag may continue on the next line, but a blank
		// line ends it: that is almost always a missing '>'.
		lineHasContent := true
		for i < limit && isSpace(src[i]) {
			if src[i] == '\n' {
				if !lineHasContent {
					return nil, &tagError{Off: start, Msg: fmt.Sprintf("tag <%s> is not closed with '>'", t.Name), Hint: "End the opening tag with '>' (or '/>' if it has no content) before a blank line."}
				}
				lineHasContent = false
			}
			i++
		}
		if i >= limit {
			return nil, &tagError{Off: start, Msg: fmt.Sprintf("tag <%s> is not closed with '>'", t.Name), Hint: "End the tag with '>' or '/>'."}
		}
		c := src[i]
		if c == '>' {
			t.End = i + 1
			return t, nil
		}
		if c == '/' {
			if i+1 < len(src) && src[i+1] == '>' {
				if t.Closing {
					return nil, &tagError{Off: i, Msg: fmt.Sprintf("invalid closing tag </%s/>", t.Name), Hint: fmt.Sprintf("Write </%s>.", t.Name)}
				}
				t.SelfClosing = true
				t.End = i + 2
				return t, nil
			}
			return nil, &tagError{Off: i, Msg: "unexpected '/' in tag", Hint: "Use '/>' at the end of a tag without content."}
		}
		if t.Closing {
			return nil, &tagError{Off: i, Msg: fmt.Sprintf("closing tag </%s> cannot have attributes", t.Name), Hint: fmt.Sprintf("Write just </%s>.", t.Name)}
		}
		if !isNameChar(c) {
			return nil, &tagError{Off: i, Msg: fmt.Sprintf("unexpected character %q in tag <%s>", c, t.Name), Hint: `Attributes look like name="value", name={literal} or just name.`}
		}
		a := lexedAttr{Off: i}
		ns := i
		for i < limit && isNameChar(src[i]) {
			i++
		}
		a.Name = string(src[ns:i])
		j := i
		for j < limit && (src[j] == ' ' || src[j] == '\t') {
			j++
		}
		if j >= limit || src[j] != '=' {
			a.Value, a.Form = true, components.FormBare
			t.Attrs = append(t.Attrs, a)
			continue
		}
		i = j + 1
		for i < limit && (src[i] == ' ' || src[i] == '\t') {
			i++
		}
		if i >= limit {
			return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("attribute '%s' has no value", a.Name), Hint: `Use name="value".`}
		}
		switch q := src[i]; q {
		case '"', '\'':
			var b strings.Builder
			i++
			closed := false
			for i < limit {
				ch := src[i]
				if ch == '\\' && i+1 < limit && (src[i+1] == q || src[i+1] == '\\') {
					b.WriteByte(src[i+1])
					i += 2
					continue
				}
				if ch == q {
					closed = true
					i++
					break
				}
				if ch == '\n' && i+1 < limit && src[i+1] == '\n' {
					break
				}
				b.WriteByte(ch)
				i++
			}
			if !closed {
				return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("the quotes of attribute '%s' are not closed", a.Name), Hint: fmt.Sprintf(`Close it with %c. To use the quote inside the value, escape it as \%c.`, q, q)}
			}
			// Like JSX, string attributes may use HTML entities (&quot;, &amp;).
			a.Value, a.Form = html.UnescapeString(b.String()), components.FormString
		case '{':
			end, ok := matchBrace(src, i, limit)
			if !ok {
				return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("the braces of attribute '%s' are not closed", a.Name), Hint: "Close it with '}'."}
			}
			raw := strings.TrimSpace(string(src[i+1 : end]))
			var v any
			if err := json.Unmarshal([]byte(raw), &v); err != nil {
				return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("value {%s} of attribute '%s' is not a literal", raw, a.Name), Hint: "JavaScript expressions are not supported. Use JSON literals: {2}, {true}, {\"text\"}, {[1, 2]}."}
			}
			if s, isStr := v.(string); isStr {
				a.Value, a.Form = s, components.FormString
			} else {
				a.Value, a.Form = v, components.FormJSON
			}
			i = end + 1
		default:
			return nil, &tagError{Off: i, Msg: fmt.Sprintf("the value of attribute '%s' needs quotes or braces", a.Name), Hint: fmt.Sprintf(`Write %s="value" or %s={2}.`, a.Name, a.Name)}
		}
		if i < limit && !isSpace(src[i]) && src[i] != '>' && src[i] != '/' {
			return nil, &tagError{Off: i, Msg: "missing space between attributes", Hint: `Separate attributes with a space: a="1" b="2".`}
		}
		t.Attrs = append(t.Attrs, a)
	}
}

// matchBrace finds the '}' matching src[open] == '{', skipping JSON strings.
func matchBrace(src []byte, open, limit int) (int, bool) {
	depth := 0
	for i := open; i < limit; i++ {
		switch src[i] {
		case '"':
			for i++; i < limit && src[i] != '"'; i++ {
				if src[i] == '\\' {
					i++
				}
			}
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}
