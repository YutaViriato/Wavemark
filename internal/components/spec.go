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
		return "text"
	case TBool:
		return "boolean"
	case TInt:
		return "integer"
	case TEnum:
		return "enum"
	case TURL:
		return "link"
	case TAsset:
		return "file"
	case TIcon:
		return "icon"
	case TDir:
		return "folder"
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
		hint := "Available components: " + strings.Join(Names(), ", ") + ". Run 'wavemark spec' for the reference."
		if alt, ok := knownAliases[c.Name]; ok {
			hint = alt
		} else if s := diag.Suggest(c.Name, Names()); s != "" {
			hint = fmt.Sprintf("Did you mean <%s>? %s", s, hint)
		}
		diags.Errorf(c.Line, c.Col, diag.CodeComponentUnknown, hint, "unknown component <%s>", c.Name)
		return nil
	}
	if spec.Inline && !inline {
		diags.Errorf(c.Line, c.Col, diag.CodeComponentSyntax, fmt.Sprintf("Write it inside text: New feature %s.", spec.Example), "<%s> is inline and must be inside a paragraph, heading or table cell", c.Name)
		return nil
	}
	if !spec.Inline && inline {
		diags.Errorf(c.Line, c.Col, diag.CodeHTMLForbidden, "Block components start on their own line, with no text before the tag.", "<%s> must be on its own line", c.Name)
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
			diags.Errorf(a.Line, a.Col, diag.CodeAttrValue, "Remove one of them.", "attribute '%s' is repeated in <%s>", a.Name, c.Name)
			continue
		}
		seen[a.Name] = true
		as := spec.attr(a.Name)
		if as == nil {
			hint := attrHint(spec)
			if alt, ok := attrAliases[spec.Name+"."+a.Name]; ok {
				hint = alt + " " + hint
			} else if s := diag.Suggest(a.Name, spec.attrNames()); s != "" {
				hint = fmt.Sprintf("Did you mean '%s'? %s", s, hint)
			}
			diags.Errorf(a.Line, a.Col, diag.CodeAttrUnknown, hint, "unknown attribute '%s' on <%s>", a.Name, c.Name)
			continue
		}
		if _, preset := spec.Preset[a.Name]; preset {
			diags.Errorf(a.Line, a.Col, diag.CodeAttrUnknown, fmt.Sprintf("<%s> already sets %s=%q; use <%s> to choose another value.", c.Name, a.Name, spec.Preset[a.Name], spec.AliasOf), "attribute '%s' cannot be used on <%s>", a.Name, c.Name)
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
			diags.Errorf(c.Line, c.Col, diag.CodeAttrMissing, fmt.Sprintf("Example: %s", spec.Example), "<%s> requires the attribute '%s'", c.Name, as.Name)
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
		diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Use the self-closing form: %s", spec.Example), "<%s> does not accept content", n.Name)
	}
	if spec.Children == ChildText && strings.TrimSpace(n.Content) == "" {
		diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Example: %s", spec.Example), "<%s> is empty", n.Name)
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
		diags.Errorf(n.Line, n.Col, diag.CodeComponentLayout, fmt.Sprintf("Set 'layout: %s' in the frontmatter or remove the component.", spec.Layouts[0]), "<%s> can only be used in the %s layout (this page uses '%s')", n.Name, strings.Join(spec.Layouts, " or "), layout)
	}

	parent := ComponentParent(n)
	if spec.Parent != "" {
		if parent == nil || canonical(parent.Name) != spec.Parent || n.Parent() != ast.Node(parent) {
			diags.Errorf(n.Line, n.Col, diag.CodeComponentParent, fmt.Sprintf("Put <%s> directly inside <%s>.", n.Name, spec.Parent), "<%s> outside <%s>", n.Name, spec.Parent)
		}
	}

	switch spec.Children {
	case ChildNone:
		if n.HasChildren() {
			diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Use the self-closing form: %s", spec.Example), "<%s> does not accept content", n.Name)
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
				diags.Errorf(line, col, diag.CodeComponentChild, fmt.Sprintf("<%s> can only contain %s.", n.Name, describeAllowed(spec.Allowed)), "content not allowed inside <%s>", n.Name)
			}
		}
		if !n.HasChildren() {
			diags.Errorf(n.Line, n.Col, diag.CodeComponentChild, fmt.Sprintf("Example:\n%s", spec.Example), "<%s> is empty", n.Name)
		}
	}
}

func describeAllowed(allowed []string) string {
	parts := make([]string, len(allowed))
	for i, a := range allowed {
		if a == ChildCode {
			parts[i] = "fenced code blocks (```)"
		} else {
			parts[i] = "<" + a + ">"
		}
	}
	return strings.Join(parts, " or ")
}

func attrHint(spec *Spec) string {
	if len(spec.Attrs) == 0 {
		return fmt.Sprintf("<%s> has no attributes.", spec.Name)
	}
	return "Valid attributes: " + strings.Join(spec.attrNames(), ", ") + "."
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
			return bad(fmt.Sprintf(`Use quotes: %s="..."`, a.Name), "'%s' on <%s> must be a quoted string", a.Name, c.Name)
		}
		s = strings.TrimSpace(s)
		if as.Required && s == "" {
			return bad("Fill in the value.", "'%s' on <%s> is empty", a.Name, c.Name)
		}
		if as.Type == TEnum && !slices.Contains(as.Enum, s) {
			hint := "Accepted values: " + strings.Join(as.Enum, ", ") + "."
			if alt, ok := valueAliases[c.Spec.Name+"."+a.Name+"="+s]; ok {
				hint = alt + " " + hint
			} else if sug := diag.Suggest(s, as.Enum); sug != "" {
				hint = fmt.Sprintf("Did you mean %q? %s", sug, hint)
			}
			return bad(hint, "invalid value %q for '%s' on <%s>", s, a.Name, c.Name)
		}
		if as.Type == TIcon {
			if _, _, ok := sketch.Lookup(s); !ok {
				hint := "Available icons: " + strings.Join(sketch.Names(), ", ") + "."
				if sug := diag.Suggest(s, sketch.AllNames()); sug != "" {
					hint = fmt.Sprintf("Did you mean %q? %s", sug, hint)
				}
				diags.Errorf(a.Line, a.Col, diag.CodeIconUnknown, hint, "unknown icon %q on <%s>", s, c.Name)
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
		return bad(fmt.Sprintf("Write just '%s' for true, or %s={false}.", a.Name, a.Name), "'%s' on <%s> must be a boolean", a.Name, c.Name)
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
			return bad(fmt.Sprintf("Use an integer: %s={2}", a.Name), "'%s' on <%s> must be an integer", a.Name, c.Name)
		}
		i := int(f)
		if (as.Min != 0 || as.Max != 0) && (i < as.Min || i > as.Max) {
			return bad("", "'%s' on <%s> must be between %d and %d (got %d)", a.Name, c.Name, as.Min, as.Max, i)
		}
		return i, true
	}
	return nil, false
}
