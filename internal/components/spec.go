package components

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// Type of an attribute value.
type Type int

const (
	TString Type = iota
	TBool
	TInt
	TEnum
	TURL   // link target: route, relative .md, #anchor or http(s)/mailto URL
	TAsset // file target: relative or absolute path to a file in the site, or https URL
)

func (t Type) String() string {
	switch t {
	case TString:
		return "texto"
	case TBool:
		return "booleano"
	case TInt:
		return "inteiro"
	case TEnum:
		return "enum"
	case TURL:
		return "link"
	case TAsset:
		return "arquivo"
	}
	return "?"
}

// AttrSpec describes one attribute of a component.
type AttrSpec struct {
	Name     string
	Type     Type
	Required bool
	Enum     []string
	Default  any
	Min, Max int // for TInt; 0,0 = unbounded
	Doc      string
}

// ChildPolicy says what a component may contain.
type ChildPolicy int

const (
	ChildMarkdown ChildPolicy = iota // any Markdown and components
	ChildNone                        // self-closing only
	ChildOnly                        // only the components listed in Allowed
)

// ChildCode is the pseudo-name for fenced code blocks in Spec.Allowed.
const ChildCode = "```"

// RenderContext carries per-page state for component rendering.
type RenderContext struct {
	// URL resolves a link or asset attribute to its final href.
	URL func(raw string) string
	// T translates interface strings (see package i18n).
	T   func(key string) string
	ids map[string]int
	// Panels maps tab-like children to their ids, assigned by the parent.
	Panels map[ast.Node]PanelID
}

// PanelID identifies one tab and its panel.
type PanelID struct {
	Tab, Panel string
	Index      int
}

// NewRenderContext creates a render context.
func NewRenderContext(url func(string) string, t func(string) string) *RenderContext {
	if url == nil {
		url = func(s string) string { return s }
	}
	if t == nil {
		t = func(s string) string { return s }
	}
	return &RenderContext{URL: url, T: t, ids: map[string]int{}, Panels: map[ast.Node]PanelID{}}
}

// NextID returns a page-unique id with the given prefix.
func (c *RenderContext) NextID(prefix string) string {
	c.ids[prefix]++
	return fmt.Sprintf("%s-%d", prefix, c.ids[prefix])
}

// RenderFunc writes the opening (entering) or closing markup of a component.
type RenderFunc func(w util.BufWriter, source []byte, n *Node, entering bool, ctx *RenderContext)

// Spec is the schema of a component.
type Spec struct {
	Name     string
	Summary  string
	Attrs    []AttrSpec
	Children ChildPolicy
	Allowed  []string // for ChildOnly
	Parent   string   // when set, must be a direct child of this component
	Layouts  []string // when set, only valid in these layouts
	Preset   map[string]any
	AliasOf  string
	Example  string
	Render   RenderFunc
}

func (s *Spec) attr(name string) *AttrSpec {
	for i := range s.Attrs {
		if s.Attrs[i].Name == name {
			return &s.Attrs[i]
		}
	}
	return nil
}

func (s *Spec) attrNames() []string {
	names := make([]string, len(s.Attrs))
	for i, a := range s.Attrs {
		names[i] = a.Name
	}
	return names
}

// LinkAttrs returns the attributes of n that point to links or assets.
func LinkAttrs(n *Node) []struct {
	Attr Attr
	Type Type
} {
	var out []struct {
		Attr Attr
		Type Type
	}
	if n.Spec == nil {
		return nil
	}
	for _, a := range n.Attrs {
		if as := n.Spec.attr(a.Name); as != nil && (as.Type == TURL || as.Type == TAsset) {
			if _, ok := a.Value.(string); ok {
				out = append(out, struct {
					Attr Attr
					Type Type
				}{a, as.Type})
			}
		}
	}
	return out
}

// Validate checks n against its schema and fills n.Spec and n.Values.
// lineOf maps a non-component node to its 1-based file line (0 if unknown).
func Validate(n *Node, layout string, lineOf func(ast.Node) int, diags *diag.List) {
	spec := Lookup(n.Name)
	if spec == nil {
		hint := "Componentes disponíveis: " + strings.Join(Names(), ", ") + ". Rode 'wavemark spec' para a referência."
		if alt, ok := knownAliases[n.Name]; ok {
			hint = alt
		} else if s := diag.Suggest(n.Name, Names()); s != "" {
			hint = fmt.Sprintf("Você quis dizer <%s>? %s", s, hint)
		}
		diags.Errorf(n.Line, n.Col, diag.CodeComponentUnknown, hint, "componente desconhecido <%s>", n.Name)
		return
	}
	n.Spec = spec
	n.Values = map[string]any{}
	for k, v := range spec.Preset {
		n.Values[k] = v
	}

	seen := map[string]bool{}
	for _, a := range n.Attrs {
		if seen[a.Name] {
			diags.Errorf(a.Line, a.Col, diag.CodeAttrValue, "Remova uma das ocorrências.", "atributo '%s' repetido em <%s>", a.Name, n.Name)
			continue
		}
		seen[a.Name] = true
		as := spec.attr(a.Name)
		if as == nil {
			hint := attrHint(spec)
			if s := diag.Suggest(a.Name, spec.attrNames()); s != "" {
				hint = fmt.Sprintf("Você quis dizer '%s'? %s", s, hint)
			}
			diags.Errorf(a.Line, a.Col, diag.CodeAttrUnknown, hint, "atributo desconhecido '%s' em <%s>", a.Name, n.Name)
			continue
		}
		if _, preset := spec.Preset[a.Name]; preset {
			diags.Errorf(a.Line, a.Col, diag.CodeAttrUnknown, fmt.Sprintf("<%s> já define %s=%q; use <%s> para escolher outro valor.", n.Name, a.Name, spec.Preset[a.Name], spec.AliasOf), "atributo '%s' não pode ser usado em <%s>", a.Name, n.Name)
			continue
		}
		if v, ok := convert(n, a, as, diags); ok {
			n.Values[a.Name] = v
		}
	}
	for _, as := range spec.Attrs {
		if _, ok := n.Values[as.Name]; ok {
			continue
		}
		if as.Required {
			diags.Errorf(n.Line, n.Col, diag.CodeAttrMissing, fmt.Sprintf("Exemplo: %s", spec.Example), "<%s> exige o atributo '%s'", n.Name, as.Name)
			continue
		}
		if as.Default != nil {
			n.Values[as.Name] = as.Default
		}
	}

	if len(spec.Layouts) > 0 && !slices.Contains(spec.Layouts, layout) {
		diags.Errorf(n.Line, n.Col, diag.CodeComponentLayout, fmt.Sprintf("Defina 'layout: %s' no frontmatter ou remova o componente.", spec.Layouts[0]), "<%s> só pode ser usado no layout %s (a página usa '%s')", n.Name, strings.Join(spec.Layouts, " ou "), layout)
	}

	parent := ComponentParent(n)
	if spec.Parent != "" {
		if parent == nil || canonical(parent.Name) != spec.Parent || n.Parent() != ast.Node(parent) {
			diags.Errorf(n.Line, n.Col, diag.CodeComponentParent, fmt.Sprintf("Coloque <%s> diretamente dentro de <%s>.", n.Name, spec.Parent), "<%s> fora de <%s>", n.Name, spec.Parent)
		}
	}

	switch spec.Children {
	case ChildNone:
		if n.HasChildren() {
			diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Use a forma autofechada: %s", spec.Example), "<%s> não aceita conteúdo", n.Name)
		}
	case ChildOnly:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			name := ""
			switch cc := c.(type) {
			case *Node:
				name = canonical(cc.Name)
			case *ast.FencedCodeBlock:
				name = ChildCode
			}
			if !slices.Contains(spec.Allowed, name) {
				line, col := n.Line, n.Col
				if cc, ok := c.(*Node); ok {
					line, col = cc.Line, cc.Col
				} else if l := lineOf(c); l > 0 {
					line, col = l, 0
				}
				diags.Errorf(line, col, diag.CodeComponentChild, fmt.Sprintf("Dentro de <%s> só pode haver %s.", n.Name, describeAllowed(spec.Allowed)), "conteúdo não permitido dentro de <%s>", n.Name)
			}
		}
		if !n.HasChildren() {
			diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Exemplo:\n%s", spec.Example), "<%s> está vazio", n.Name)
		}
	}
}

func describeAllowed(allowed []string) string {
	parts := make([]string, len(allowed))
	for i, a := range allowed {
		if a == ChildCode {
			parts[i] = "blocos de código cercados (```)"
		} else {
			parts[i] = "<" + a + ">"
		}
	}
	return strings.Join(parts, " ou ")
}

func attrHint(spec *Spec) string {
	if len(spec.Attrs) == 0 {
		return fmt.Sprintf("<%s> não tem atributos.", spec.Name)
	}
	return "Atributos válidos: " + strings.Join(spec.attrNames(), ", ") + "."
}

func convert(n *Node, a Attr, as *AttrSpec, diags *diag.List) (any, bool) {
	bad := func(hint, format string, args ...any) (any, bool) {
		diags.Errorf(a.Line, a.Col, diag.CodeAttrValue, hint, format, args...)
		return nil, false
	}
	switch as.Type {
	case TString, TURL, TAsset, TEnum:
		s, ok := a.Value.(string)
		if !ok || a.Form != FormString {
			return bad(fmt.Sprintf(`Use aspas: %s="..."`, a.Name), "'%s' em <%s> precisa ser um texto entre aspas", a.Name, n.Name)
		}
		s = strings.TrimSpace(s)
		if as.Required && s == "" {
			return bad("Preencha o valor.", "'%s' em <%s> está vazio", a.Name, n.Name)
		}
		if as.Type == TEnum && !slices.Contains(as.Enum, s) {
			hint := "Valores aceitos: " + strings.Join(as.Enum, ", ") + "."
			if sug := diag.Suggest(s, as.Enum); sug != "" {
				hint = fmt.Sprintf("Você quis dizer %q? %s", sug, hint)
			}
			return bad(hint, "valor %q inválido para '%s' em <%s>", s, a.Name, n.Name)
		}
		return s, true
	case TBool:
		if b, ok := a.Value.(bool); ok {
			return b, true
		}
		return bad(fmt.Sprintf("Escreva só '%s' para verdadeiro, ou %s={false}.", a.Name, a.Name), "'%s' em <%s> precisa ser booleano", a.Name, n.Name)
	case TInt:
		f, ok := a.Value.(float64)
		if !ok || a.Form != FormJSON || f != math.Trunc(f) {
			return bad(fmt.Sprintf("Use chaves: %s={2}", a.Name), "'%s' em <%s> precisa ser um inteiro entre chaves", a.Name, n.Name)
		}
		i := int(f)
		if (as.Min != 0 || as.Max != 0) && (i < as.Min || i > as.Max) {
			return bad("", "'%s' em <%s> precisa estar entre %d e %d (recebeu %d)", a.Name, n.Name, as.Min, as.Max, i)
		}
		return i, true
	}
	return nil, false
}
