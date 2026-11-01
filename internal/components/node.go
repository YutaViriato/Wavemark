// Package components defines Wavemark's declarative components: their AST
// node, their schemas (attributes, children, parents), validation and HTML
// rendering.
package components

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
)

// KindComponent is the goldmark node kind of every component.
var KindComponent = ast.NewNodeKind("Component")

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

// Node is a component in the Markdown AST.
type Node struct {
	ast.BaseBlock
	Name        string
	Attrs       []Attr
	Line, Col   int // position of the opening tag, 1-based in the page file
	Closed      bool
	SelfClosing bool
	Spec        *Spec          // nil when the component is unknown
	Values      map[string]any // validated values, defaults applied
}

// NewNode creates a component node.
func NewNode(name string, line, col int) *Node {
	return &Node{Name: name, Line: line, Col: col}
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

// Attr returns the attribute with the given name.
func (n *Node) Attr(name string) (Attr, bool) {
	for _, a := range n.Attrs {
		if a.Name == name {
			return a, true
		}
	}
	return Attr{}, false
}

// Str returns a validated string value ("" when absent).
func (n *Node) Str(name string) string {
	s, _ := n.Values[name].(string)
	return s
}

// Bool returns a validated boolean value.
func (n *Node) Bool(name string) bool {
	b, _ := n.Values[name].(bool)
	return b
}

// Int returns a validated integer value.
func (n *Node) Int(name string) int {
	i, _ := n.Values[name].(int)
	return i
}

// ComponentParent returns the nearest enclosing component, or nil.
func ComponentParent(node ast.Node) *Node {
	for p := node.Parent(); p != nil; p = p.Parent() {
		if c, ok := p.(*Node); ok {
			return c
		}
	}
	return nil
}
