// Package components defines Wavemark's declarative components: their AST
// nodes, their schemas (attributes, children, parents), validation and HTML
// rendering.
package components

import (
	"fmt"
	"time"

	"github.com/yuin/goldmark/ast"
)

// KindComponent is the goldmark node kind of block components.
var KindComponent = ast.NewNodeKind("Component")

// KindInline is the goldmark node kind of inline components (Badge, Icon).
var KindInline = ast.NewNodeKind("InlineComponent")

// AttrForm is how an attribute value was written in the source.
type AttrForm int

const (
	FormString AttrForm = iota // name="value" or name='value'
	FormJSON                   // name={literal}
	FormBare                   // name
)

// Attr is an attribute as written in the source.
type Attr struct {
	Name  string
	Value any // string, bool, float64, []any or map[string]any
	Form  AttrForm
	Line  int // 1-based, in the page file
	Col   int
}

// Core holds what block and inline components share.
type Core struct {
	Name      string
	Attrs     []Attr
	Line, Col int            // position of the opening tag, 1-based in the page file
	Spec      *Spec          // nil when the component is unknown
	Values    map[string]any // validated values, defaults applied
}

// Attr returns the attribute with the given name.
func (c *Core) Attr(name string) (Attr, bool) {
	for _, a := range c.Attrs {
		if a.Name == name {
			return a, true
		}
	}
	return Attr{}, false
}

// Str returns a validated string value ("" when absent).
func (c *Core) Str(name string) string {
	s, _ := c.Values[name].(string)
	return s
}

// Bool returns a validated boolean value.
func (c *Core) Bool(name string) bool {
	b, _ := c.Values[name].(bool)
	return b
}

// Int returns a validated integer value.
func (c *Core) Int(name string) int {
	i, _ := c.Values[name].(int)
	return i
}

// Node is a block component in the Markdown AST.
type Node struct {
	ast.BaseBlock
	Core
	Closed      bool
	SelfClosing bool
}

// NewNode creates a block component node.
func NewNode(name string, line, col int) *Node {
	return &Node{Core: Core{Name: name, Line: line, Col: col}}
}

// Kind implements ast.Node.
func (n *Node) Kind() ast.NodeKind { return KindComponent }

// Dump implements ast.Node.
func (n *Node) Dump(source []byte, level int) {
	m := map[string]string{"Name": n.Name}
	for _, a := range n.Attrs {
		m["@"+a.Name] = fmt.Sprint(a.Value)
	}
	ast.DumpHelper(n, source, level, m, nil)
}

// InlineNode is an inline component inside a paragraph or heading.
type InlineNode struct {
	ast.BaseInline
	Core
	Content string // text between the tags
}

// NewInline creates an inline component node.
func NewInline(name string, line, col int) *InlineNode {
	return &InlineNode{Core: Core{Name: name, Line: line, Col: col}}
}

// Kind implements ast.Node.
func (n *InlineNode) Kind() ast.NodeKind { return KindInline }

// Dump implements ast.Node.
func (n *InlineNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Name": n.Name, "Content": n.Content}, nil)
}

// ComponentParent returns the nearest enclosing block component, or nil.
func ComponentParent(node ast.Node) *Node {
	for p := node.Parent(); p != nil; p = p.Parent() {
		if c, ok := p.(*Node); ok {
			return c
		}
	}
	return nil
}

// PageInfo describes a page for components that list pages.
type PageInfo struct {
	Title       string
	Description string
	Route       string
	Date        time.Time
	Author      string
	Icon        string
}
