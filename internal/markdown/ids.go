package markdown

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
)

// accents folds Latin letters with diacritics to ASCII so that Portuguese,
// Spanish and French headings get readable anchors ("Configuração" →
// "configuracao") instead of goldmark's default, which drops the letter.
var accents = map[rune]string{
	'á': "a", 'à': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a", 'ā': "a",
	'é': "e", 'è': "e", 'ê': "e", 'ë': "e", 'ē': "e",
	'í': "i", 'ì': "i", 'î': "i", 'ï': "i", 'ī': "i",
	'ó': "o", 'ò': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o", 'ō': "o",
	'ú': "u", 'ù': "u", 'û': "u", 'ü': "u", 'ū': "u",
	'ç': "c", 'ñ': "n", 'ý': "y", 'ÿ': "y", 'ß': "ss", 'æ': "ae", 'œ': "oe",
}

// Slug converts heading text to an anchor id.
func Slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if a, ok := accents[r]; ok {
			b.WriteString(a)
			dash = false
			continue
		}
		switch {
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			b.WriteRune(r)
			dash = false
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r) // keep non-Latin scripts as-is
			dash = false
		case r == ' ' || r == '-' || r == '_' || r == '/' || r == '.':
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// ids implements parser.IDs with Slug and per-document uniqueness.
type ids struct{ used map[string]bool }

func newIDs() *ids { return &ids{used: map[string]bool{}} }

func (s *ids) Generate(value []byte, kind ast.NodeKind) []byte {
	base := Slug(stripInlineMarkup(string(value)))
	if base == "" {
		base = "secao"
	}
	id := base
	for i := 1; s.used[id]; i++ {
		id = base + "-" + strconv.Itoa(i)
	}
	s.used[id] = true
	return []byte(id)
}

func (s *ids) Put(value []byte) { s.used[string(value)] = true }

var inlineComponentRe = regexp.MustCompile(`<[A-Z][A-Za-z]*\b[^>]*/>|<([A-Z][A-Za-z]*)\b[^>]*>[^<]*</[A-Z][A-Za-z]*>`)

// stripInlineMarkup removes link targets and inline components from raw
// heading text so that "[Guia](./guia.md) <Badge>Novo</Badge>" yields "Guia".
func stripInlineMarkup(s string) string {
	s = inlineComponentRe.ReplaceAllString(s, "")
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == ']' && i+1 < len(s) && s[i+1] == '(' {
			if j := strings.IndexByte(s[i:], ')'); j > 0 {
				i += j
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
