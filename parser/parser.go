package parser

import (
	"fmt"
	"github.com/iota-uz/margo/lexer"
	"github.com/yuin/goldmark/ast"
	"strings"
)

type Parser struct {
	lexer  *lexer.Lexer
	indent int
}

func NewMargoParser(content string) *Parser {
	return &Parser{
		lexer:  lexer.New(content),
		indent: 0,
	}
}

func (p *Parser) Parse() ([]Node, error) {
	return p.parseBlock()
}

func (p *Parser) parseBlock() ([]Node, error) {
	var nodes []Node
	baseIndent := p.indent
	p.indent = 0
	for token := p.lexer.Peek(); token.Type != lexer.EOF; token = p.lexer.Peek() {
		if token.Type == lexer.Indent {
			p.indent++
			p.lexer.Next()
			continue
		}
		if token.Type == lexer.LineBreak {
			p.indent = 0
			p.lexer.Next()
			continue
		}
		if p.indent <= baseIndent && baseIndent != 0 {
			break
		}
		switch token.Type {
		case lexer.Component:
			node, err := p.parseComponentNode()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		case lexer.Text:
			p.lexer.Next()
			nodes = append(nodes, &TextNode{Value: token.Value})
		default:
			panic(fmt.Sprintf("unexpected %v: %v", token.Type, token.Value))
		}
	}
	return nodes, nil
}

func (p *Parser) parseString() (string, error) {
	if token := p.lexer.Next(); token.Type != lexer.Quote {
		return "", fmt.Errorf(
			"%d:%d expected opening quote got %v",
			token.Line,
			token.Column,
			token.Value,
		)
	}
	token := p.lexer.Next()
	if token.Type != lexer.Text {
		return "", fmt.Errorf(
			"malformed string at %d:%d",
			token.Line,
			token.Column,
		)
	}
	if next := p.lexer.Next(); next.Type != lexer.Quote {
		return "", fmt.Errorf(
			"%d:%d expected closing quote got %v",
			next.Line,
			next.Column,
			next.Value,
		)
	}
	return token.Value, nil
}

func (p *Parser) parseMultiLineString(currentIndent int) (string, error) {
	var s string
	indent := currentIndent
	for token := p.lexer.Next(); token.Type != lexer.EOF; token = p.lexer.Next() {
		if token.Type == lexer.LineBreak {
			indent = 0
			continue
		}
		switch token.Type {
		case lexer.Indent:
			indent++
			continue
		case lexer.Text:
			s += token.Value
			continue
		default:
			// do nothing
		}

		if indent <= currentIndent {
			return "", fmt.Errorf(
				"%d:%d expected indent greater than %d got %d",
				token.Column,
				token.Type,
				currentIndent,
				indent,
			)
		}
		if token.Type != lexer.Text {
			return "", fmt.Errorf(
				"%d:%d expected string got %v",
				token.Column,
				token.Type,
				token.Value,
			)
		}
		s += token.Value
	}
	return s, nil
}

func (p *Parser) parsePropertyValue() (any, error) {
	next := p.lexer.Peek()
	if next.Type == lexer.Quote {
		return p.parseString()
	}
	if next.Type == lexer.LineBreak {
		p.lexer.Next()
		nodes, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		componentNodes := 0
		for _, node := range nodes {
			if _, ok := node.(*ComponentNode); ok {
				componentNodes++
			}
		}
		if componentNodes > 1 {
			return nil, fmt.Errorf(
				"%d:%d a property value can only contain one component got %d",
				next.Line,
				next.Column,
				componentNodes,
			)
		}
		var text []string
		if componentNodes == 0 {
			for _, node := range nodes {
				text = append(text, node.(*TextNode).Value)
			}
			return strings.Join(text, "\n"), nil
		}
		return nodes[0], nil

	}
	return nil, fmt.Errorf(
		"%d:%d expected \" or newline got %v",
		next.Line,
		next.Column,
		next.Value,
	)
}

func (p *Parser) parseProperty() (string, any, error) {
	name := p.lexer.Next().Value
	if next := p.lexer.Next(); next.Type != lexer.Colon {
		return "", nil, fmt.Errorf(
			"%d:%d expected colon after property name got %v",
			next.Line,
			next.Column,
			next.Value,
		)
	}
	v, err := p.parsePropertyValue()
	if err != nil {
		return "", nil, err
	}
	return name, v, nil
}

func (p *Parser) parseComponentNode() (*ComponentNode, error) {
	node := &ComponentNode{
		Name:       p.lexer.Next().Value,
		attributes: make([]ast.Attribute, 0),
	}
	baseIndent := p.indent

	for token := p.lexer.Peek(); token.Type != lexer.EOF; token = p.lexer.Peek() {
		if token.Type == lexer.Indent {
			p.indent++
			p.lexer.Next()
			continue
		}
		if token.Type == lexer.LineBreak {
			p.indent = 0
			p.lexer.Next()
			continue
		}
		if p.indent <= baseIndent {
			break
		}
		switch token.Type {
		case lexer.Text:
			p.lexer.Next()
			node.children = append(node.children, &TextNode{Value: token.Value})
		case lexer.Component:
			child, err := p.parseComponentNode()
			if err != nil {
				return nil, err
			}
			node.children = append(node.children, child)
		case lexer.Bool:
			p.lexer.Next()
			node.attributes = append(node.attributes, ast.Attribute{
				Name:  []byte(token.Value),
				Value: true,
			})
		case lexer.Property:
			k, v, err := p.parseProperty()
			if err != nil {
				return nil, err
			}
			node.attributes = append(node.attributes, ast.Attribute{
				Name:  []byte(k),
				Value: v,
			})
		default:
			panic(fmt.Sprintf("unexpected %v: %v", token.Type, token.Value))
		}
	}
	return node, nil
}
