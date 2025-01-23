package parser

import "github.com/yuin/goldmark/ast"

type Props map[string]any

type Node interface {
	Children() []Node
}

type ComponentNode struct {
	attributes []ast.Attribute
	Name       string
	children   []Node
}

func (n *ComponentNode) Attributes() []ast.Attribute {
	return n.attributes
}

func (n *ComponentNode) Children() []Node {
	return n.children
}

type TextNode struct {
	Value string
}

func (t *TextNode) Children() []Node { return nil }

var KindMargoNode = ast.NewNodeKind("MargoNode")

type Document struct {
	ast.BaseBlock
	Children []Node
}

// Kind implements Node.Kind.
func (n *Document) Kind() ast.NodeKind {
	return KindMargoNode
}

// Dump implements Node.Dump
func (n *Document) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
