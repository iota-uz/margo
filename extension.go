package margo

import (
	"context"
	"io"
	"log"
	"reflect"

	"github.com/a-h/templ"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	margoparser "github.com/iota-uz/margo/parser"
	"github.com/iota-uz/margo/registry"
	margorender "github.com/iota-uz/margo/renderer"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

type Markdown interface {
	goldmark.Markdown
	ConvertToTempl(source []byte, opts ...parser.ParseOption) templ.Component
}

type markdown struct {
	parser     parser.Parser
	renderer   margorender.Renderer
	extensions []goldmark.Extender
}

// New returns a new Markdown with given options.
func New(reg registry.Layout) Markdown {
	defaultParser := goldmark.DefaultParser()
	defaultParser.AddOptions(
		parser.WithAutoHeadingID(),
		parser.WithAttribute(),
	)
	md := &markdown{
		parser:   defaultParser,
		renderer: margorender.NewRenderer(reg, renderer.WithNodeRenderers(util.Prioritized(html.NewRenderer(), 1000))),
		extensions: []goldmark.Extender{
			Extension(reg),
			meta.Meta,
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("xcode-dark"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
		},
	}
	for _, e := range md.extensions {
		e.Extend(md)
	}
	return md
}

func (m *markdown) Convert(source []byte, writer io.Writer, opts ...parser.ParseOption) error {
	reader := text.NewReader(source)
	doc := m.parser.Parse(reader, opts...)
	return m.renderer.Render(writer, source, doc)
}

func (m *markdown) ConvertToTempl(source []byte, opts ...parser.ParseOption) templ.Component {
	reader := text.NewReader(source)
	doc := m.parser.Parse(reader, opts...)
	return m.renderer.RenderToTempl(source, doc)
}

func (m *markdown) Parser() parser.Parser {
	return m.parser
}

func (m *markdown) SetParser(v parser.Parser) {
	m.parser = v
}

func (m *markdown) Renderer() renderer.Renderer {
	return m.renderer
}

func (m *markdown) SetRenderer(v renderer.Renderer) {
	// TODO: deal with this bullshit
	log.Println("SetRenderer", reflect.TypeOf(v))
	//m.renderer = v
}

func Extension(layout registry.Layout) goldmark.Extender {
	return &extender{
		layout: layout,
	}
}

type extender struct {
	layout registry.Layout
	ctx    context.Context
}

func (e *extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(margoparser.BlockParser(), 10),
		),
	)
	m.Renderer().AddOptions(
		html.WithUnsafe(),
	)
}
