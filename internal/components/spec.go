package components

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/YutaViriato/Wavemark/internal/diag"
	"github.com/YutaViriato/Wavemark/internal/sketch"
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
	TIcon  // name of a built-in hand-drawn icon
	TDir   // folder of the site, relative to the page or absolute
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
	case TIcon:
		return "ícone"
	case TDir:
		return "pasta"
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
	ChildText                        // inline components: plain text between the tags
)

// ChildCode is the pseudo-name for fenced code blocks in Spec.Allowed.
const ChildCode = "```"

// RenderContext carries per-page state for component rendering.
type RenderContext struct {
	// URL resolves a link or asset attribute to its final href.
	URL func(raw string) string
	// T translates interface strings (see package i18n).
	T func(key string) string
	// Pages lists the pages of a folder (for PageList).
	Pages func(dir string) []PageInfo
	// Layout and language of the page being rendered.
	Layout string
	Lang   string
	ids    map[string]int
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
	return &RenderContext{URL: url, T: t, ids: map[string]int{}, Panels: map[ast.Node]PanelID{}, Pages: func(string) []PageInfo { return nil }}
}

// NextID returns a page-unique id with the given prefix.
func (c *RenderContext) NextID(prefix string) string {
	c.ids[prefix]++
	return fmt.Sprintf("%s-%d", prefix, c.ids[prefix])
}

// RenderFunc writes the opening (entering) or closing markup of a block
// component.
type RenderFunc func(w util.BufWriter, source []byte, n *Node, entering bool, ctx *RenderContext)

// InlineRenderFunc writes an inline component.
type InlineRenderFunc func(w util.BufWriter, n *InlineNode, ctx *RenderContext)

// Spec is the schema of a component.
type Spec struct {
	Name         string
	Summary      string
	Attrs        []AttrSpec
	Children     ChildPolicy
	Allowed      []string // for ChildOnly
	Parent       string   // when set, must be a direct child of this component
	Layouts      []string // when set, only valid in these layouts
	Preset       map[string]any
	AliasOf      string
	Inline       bool // written inside a paragraph: <Badge>New</Badge>
	SkipChildren bool // the renderer consumes the children itself
	Example      string
	Render       RenderFunc
	RenderInline InlineRenderFunc
	check        func(c *Core, diags *diag.List)
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

// TypedAttr is an attribute whose value must be resolved by the site.
type TypedAttr struct {
	Attr Attr
	Type Type
}

// LinkAttrs returns the attributes of c that point to links, assets or
// folders.
func LinkAttrs(c *Core) []TypedAttr {
	var out []TypedAttr
	if c.Spec == nil {
		return nil
	}
	for _, a := range c.Attrs {
		if as := c.Spec.attr(a.Name); as != nil && (as.Type == TURL || as.Type == TAsset || as.Type == TDir) {
			if _, ok := a.Value.(string); ok {
				out = append(out, TypedAttr{a, as.Type})
			}
		}
	}
	return out
}

// lookupFor finds the schema of a block or inline component and reports
// components used in the wrong position.
func lookupFor(c *Core, inline bool, diags *diag.List) *Spec {
	spec := Lookup(c.Name)
	if spec == nil {
		hint := "Componentes disponíveis: " + strings.Join(Names(), ", ") + ". Rode 'wavemark spec' para a referência."
		if alt, ok := knownAliases[c.Name]; ok {
			hint = alt
		} else if s := diag.Suggest(c.Name, Names()); s != "" {
			hint = fmt.Sprintf("Você quis dizer <%s>? %s", s, hint)
		}
		diags.Errorf(c.Line, c.Col, diag.CodeComponentUnknown, hint, "componente desconhecido <%s>", c.Name)
		return nil
	}
	if spec.Inline && !inline {
		diags.Errorf(c.Line, c.Col, diag.CodeComponentSyntax, fmt.Sprintf("Escreva no meio do texto: Recurso novo %s.", spec.Example), "<%s> é inline e precisa ficar dentro de um parágrafo, título ou célula", c.Name)
		return nil
	}
	if !spec.Inline && inline {
		diags.Errorf(c.Line, c.Col, diag.CodeHTMLForbidden, "Componentes de bloco começam em uma linha própria, sem texto antes da tag.", "<%s> precisa ficar em uma linha própria", c.Name)
		return nil
	}
	return spec
}

// validateAttrs checks attributes against the schema and fills c.Values.
func validateAttrs(c *Core, spec *Spec, diags *diag.List) {
	c.Spec = spec
	c.Values = map[string]any{}
	for k, v := range spec.Preset {
		c.Values[k] = v
	}
	seen := map[string]bool{}
	for _, a := range c.Attrs {
		if seen[a.Name] {
			diags.Errorf(a.Line, a.Col, diag.CodeAttrValue, "Remova uma das ocorrências.", "atributo '%s' repetido em <%s>", a.Name, c.Name)
			continue
		}
		seen[a.Name] = true
		as := spec.attr(a.Name)
		if as == nil {
			hint := attrHint(spec)
			if alt, ok := attrAliases[spec.Name+"."+a.Name]; ok {
				hint = alt + " " + hint
			} else if s := diag.Suggest(a.Name, spec.attrNames()); s != "" {
				hint = fmt.Sprintf("Você quis dizer '%s'? %s", s, hint)
			}
			diags.Errorf(a.Line, a.Col, diag.CodeAttrUnknown, hint, "atributo desconhecido '%s' em <%s>", a.Name, c.Name)
			continue
		}
		if _, preset := spec.Preset[a.Name]; preset {
			diags.Errorf(a.Line, a.Col, diag.CodeAttrUnknown, fmt.Sprintf("<%s> já define %s=%q; use <%s> para escolher outro valor.", c.Name, a.Name, spec.Preset[a.Name], spec.AliasOf), "atributo '%s' não pode ser usado em <%s>", a.Name, c.Name)
			continue
		}
		if v, ok := convert(c, a, as, diags); ok {
			c.Values[a.Name] = v
		}
	}
	for _, as := range spec.Attrs {
		if _, ok := c.Values[as.Name]; ok {
			continue
		}
		if as.Required {
			diags.Errorf(c.Line, c.Col, diag.CodeAttrMissing, fmt.Sprintf("Exemplo: %s", spec.Example), "<%s> exige o atributo '%s'", c.Name, as.Name)
			continue
		}
		if as.Default != nil {
			c.Values[as.Name] = as.Default
		}
	}
}

// ValidateInline checks an inline component.
func ValidateInline(n *InlineNode, diags *diag.List) {
	spec := lookupFor(&n.Core, true, diags)
	if spec == nil {
		return
	}
	validateAttrs(&n.Core, spec, diags)
	if spec.Children == ChildNone && strings.TrimSpace(n.Content) != "" {
		diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Use a forma autofechada: %s", spec.Example), "<%s> não aceita conteúdo", n.Name)
	}
	if spec.Children == ChildText && strings.TrimSpace(n.Content) == "" {
		diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Exemplo: %s", spec.Example), "<%s> está vazio", n.Name)
	}
}

// Validate checks n against its schema and fills n.Spec and n.Values.
// lineOf maps a non-component node to its 1-based file line (0 if unknown).
func Validate(n *Node, layout string, lineOf func(ast.Node) int, diags *diag.List) {
	spec := lookupFor(&n.Core, false, diags)
	if spec == nil {
		return
	}
	validateAttrs(&n.Core, spec, diags)

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

func convert(c *Core, a Attr, as *AttrSpec, diags *diag.List) (any, bool) {
	bad := func(hint, format string, args ...any) (any, bool) {
		diags.Errorf(a.Line, a.Col, diag.CodeAttrValue, hint, format, args...)
		return nil, false
	}
	switch as.Type {
	case TString, TURL, TAsset, TEnum, TIcon, TDir:
		s, ok := a.Value.(string)
		if !ok || a.Form != FormString {
			return bad(fmt.Sprintf(`Use aspas: %s="..."`, a.Name), "'%s' em <%s> precisa ser um texto entre aspas", a.Name, c.Name)
		}
		s = strings.TrimSpace(s)
		if as.Required && s == "" {
			return bad("Preencha o valor.", "'%s' em <%s> está vazio", a.Name, c.Name)
		}
		if as.Type == TEnum && !slices.Contains(as.Enum, s) {
			hint := "Valores aceitos: " + strings.Join(as.Enum, ", ") + "."
			if alt, ok := valueAliases[c.Spec.Name+"."+a.Name+"="+s]; ok {
				hint = alt + " " + hint
			} else if sug := diag.Suggest(s, as.Enum); sug != "" {
				hint = fmt.Sprintf("Você quis dizer %q? %s", sug, hint)
			}
			return bad(hint, "valor %q inválido para '%s' em <%s>", s, a.Name, c.Name)
		}
		if as.Type == TIcon {
			if _, _, ok := sketch.Lookup(s); !ok {
				hint := "Ícones disponíveis: " + strings.Join(sketch.Names(), ", ") + "."
				if sug := diag.Suggest(s, sketch.AllNames()); sug != "" {
					hint = fmt.Sprintf("Você quis dizer %q? %s", sug, hint)
				}
				diags.Errorf(a.Line, a.Col, diag.CodeIconUnknown, hint, "ícone desconhecido %q em <%s>", s, c.Name)
				return nil, false
			}
		}
		return s, true
	case TBool:
		if b, ok := a.Value.(bool); ok {
			return b, true
		}
		if s, ok := a.Value.(string); ok && (s == "true" || s == "false") {
			return s == "true", true // open="true" is unambiguous enough
		}
		return bad(fmt.Sprintf("Escreva só '%s' para verdadeiro, ou %s={false}.", a.Name, a.Name), "'%s' em <%s> precisa ser booleano", a.Name, c.Name)
	case TInt:
		f, ok := a.Value.(float64)
		if s, isStr := a.Value.(string); isStr {
			// cols="3" is a common habit; accept digits written as text.
			var n int
			if _, err := fmt.Sscanf(s, "%d", &n); err == nil && fmt.Sprint(n) == strings.TrimSpace(s) {
				f, ok = float64(n), true
			}
		}
		if !ok || f != math.Trunc(f) {
			return bad(fmt.Sprintf("Use um número inteiro: %s={2}", a.Name), "'%s' em <%s> precisa ser um inteiro", a.Name, c.Name)
		}
		i := int(f)
		if (as.Min != 0 || as.Max != 0) && (i < as.Min || i > as.Max) {
			return bad("", "'%s' em <%s> precisa estar entre %d e %d (recebeu %d)", a.Name, c.Name, as.Min, as.Max, i)
		}
		return i, true
	}
	return nil, false
}
