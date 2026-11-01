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
		return nil, &tagError{Off: nameStart, Msg: fmt.Sprintf("nome de componente inválido <%s>", t.Name), Hint: "Nomes de componente são palavras em PascalCase, como <Card>."}
	}
	limit := min(len(src), start+maxTagBytes)
	for {
		// Skip whitespace. A tag may continue on the next line, but a blank
		// line ends it: that is almost always a missing '>'.
		lineHasContent := true
		for i < limit && isSpace(src[i]) {
			if src[i] == '\n' {
				if !lineHasContent {
					return nil, &tagError{Off: start, Msg: fmt.Sprintf("tag <%s> não foi fechada com '>'", t.Name), Hint: "Termine a tag de abertura com '>' (ou '/>' se não tiver conteúdo) antes de uma linha em branco."}
				}
				lineHasContent = false
			}
			i++
		}
		if i >= limit {
			return nil, &tagError{Off: start, Msg: fmt.Sprintf("tag <%s> não foi fechada com '>'", t.Name), Hint: "Termine a tag com '>' ou '/>'."}
		}
		c := src[i]
		if c == '>' {
			t.End = i + 1
			return t, nil
		}
		if c == '/' {
			if i+1 < len(src) && src[i+1] == '>' {
				if t.Closing {
					return nil, &tagError{Off: i, Msg: fmt.Sprintf("tag de fechamento </%s/> inválida", t.Name), Hint: fmt.Sprintf("Escreva </%s>.", t.Name)}
				}
				t.SelfClosing = true
				t.End = i + 2
				return t, nil
			}
			return nil, &tagError{Off: i, Msg: "'/' inesperado na tag", Hint: "Use '/>' no fim de uma tag sem conteúdo."}
		}
		if t.Closing {
			return nil, &tagError{Off: i, Msg: fmt.Sprintf("tag de fechamento </%s> não aceita atributos", t.Name), Hint: fmt.Sprintf("Escreva apenas </%s>.", t.Name)}
		}
		if !isNameChar(c) {
			return nil, &tagError{Off: i, Msg: fmt.Sprintf("caractere inesperado %q na tag <%s>", c, t.Name), Hint: `Atributos têm a forma nome="valor", nome={literal} ou só nome.`}
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
			return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("atributo '%s' sem valor", a.Name), Hint: `Use nome="valor".`}
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
				return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("aspas do atributo '%s' não foram fechadas", a.Name), Hint: fmt.Sprintf(`Feche com %c. Para usar a aspa dentro do valor, escape com \%c.`, q, q)}
			}
			// Like JSX, string attributes may use HTML entities (&quot;, &amp;).
			a.Value, a.Form = html.UnescapeString(b.String()), components.FormString
		case '{':
			end, ok := matchBrace(src, i, limit)
			if !ok {
				return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("chaves do atributo '%s' não foram fechadas", a.Name), Hint: "Feche com '}'."}
			}
			raw := strings.TrimSpace(string(src[i+1 : end]))
			var v any
			if err := json.Unmarshal([]byte(raw), &v); err != nil {
				return nil, &tagError{Off: a.Off, Msg: fmt.Sprintf("valor {%s} do atributo '%s' não é um literal", raw, a.Name), Hint: "Expressões JavaScript não são suportadas. Use literais JSON: {2}, {true}, {\"texto\"}, {[1, 2]}."}
			}
			if s, isStr := v.(string); isStr {
				a.Value, a.Form = s, components.FormString
			} else {
				a.Value, a.Form = v, components.FormJSON
			}
			i = end + 1
		default:
			return nil, &tagError{Off: i, Msg: fmt.Sprintf("valor do atributo '%s' precisa de aspas ou chaves", a.Name), Hint: fmt.Sprintf(`Escreva %s="valor" ou %s={2}.`, a.Name, a.Name)}
		}
		if i < limit && !isSpace(src[i]) && src[i] != '>' && src[i] != '/' {
			return nil, &tagError{Off: i, Msg: "falta espaço entre atributos", Hint: `Separe os atributos com espaço: a="1" b="2".`}
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
