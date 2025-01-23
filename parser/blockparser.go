package parser

import (
	"bytes"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func BlockParser() parser.BlockParser {
	return &blockParser{}
}

type blockParser struct{}

func (m *blockParser) Trigger() []byte {
	return []byte("```margo")
}

func (m *blockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if !bytes.HasPrefix(line, m.Trigger()) {
		return nil, parser.NoChildren
	}
	node := &Document{}
	return node, parser.NoChildren
}

func (m *blockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()

	if bytes.HasPrefix(line, []byte("```")) {
		reader.Advance(segment.Len())
		return parser.Close
	}
	node.Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (m *blockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	lines := node.Lines()
	var buf bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		buf.Write(segment.Value(reader.Source()))
	}
	nodes, err := NewMargoParser(string(buf.Bytes())).Parse()
	if err != nil {
		panic(err)
	}
	n := node.(*Document)
	n.Children = nodes
}

func (m *blockParser) CanInterruptParagraph() bool {
	return true
}

func (m *blockParser) CanAcceptIndentedLine() bool {
	return true
}
