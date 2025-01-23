package renderer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/a-h/templ"
	"github.com/iota-uz/margo"
	"github.com/iota-uz/margo/parser"
	"github.com/iota-uz/margo/registry"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"io"
	"sync"
)

type Renderer interface {
	renderer.Renderer
	RenderToTempl(source []byte, n ast.Node) templ.Component
}

// ContextKey is used for context value storage
type ContextKey struct{}

var slotKey = ContextKey{}
var layoutKey = ContextKey{}

// WithLayout adds a layout to context
func WithLayout(ctx context.Context, layout string) context.Context {
	return context.WithValue(ctx, layoutKey, layout)
}

// GetLayout retrieves a layout from context
func GetLayout(ctx context.Context) (string, bool) {
	layout, ok := ctx.Value(layoutKey).(string)
	return layout, ok
}

// WithSlot adds a component to context
func WithSlot(ctx context.Context, component templ.Component) context.Context {
	return context.WithValue(ctx, slotKey, component)
}

// GetSlot retrieves a component from context
func GetSlot(ctx context.Context) (templ.Component, bool) {
	c, ok := ctx.Value(slotKey).(templ.Component)
	return c, ok
}

// MarkdownRenderer implements custom markdown rendering
type MarkdownRenderer struct {
	layout               registry.Layout
	config               *renderer.Config
	options              map[renderer.OptionName]interface{}
	nodeRendererFuncsTmp map[ast.NodeKind]renderer.NodeRendererFunc
	nodeRendererFuncs    []renderer.NodeRendererFunc
	maxKind              int
	initSync             sync.Once
	componentBuilder     *ComponentBuilder
	nodeRenderer         *NodeRenderer
}

func (r *MarkdownRenderer) RegisterFuncs(registerer renderer.NodeRendererFuncRegisterer) {
	// do nothing
}

// NewRenderer creates a new markdown renderer
func NewRenderer(reg registry.Layout, options ...renderer.Option) margo.Renderer {
	config := renderer.NewConfig()
	for _, opt := range options {
		opt.SetConfig(config)
	}

	builder := NewComponentBuilder(reg)
	r := &MarkdownRenderer{
		layout:               reg,
		config:               config,
		options:              make(map[renderer.OptionName]interface{}),
		nodeRendererFuncsTmp: make(map[ast.NodeKind]renderer.NodeRendererFunc),
		componentBuilder:     builder,
	}
	r.nodeRenderer = NewNodeRenderer(builder, reg, r)
	return r
}

func (r *MarkdownRenderer) AddOptions(opts ...renderer.Option) {
	for _, opt := range opts {
		opt.SetConfig(r.config)
	}
}

func (r *MarkdownRenderer) Register(kind ast.NodeKind, v renderer.NodeRendererFunc) {
	r.nodeRendererFuncsTmp[kind] = v
	if int(kind) > r.maxKind {
		r.maxKind = int(kind)
	}
}

func (r *MarkdownRenderer) RenderToTempl(source []byte, n ast.Node) templ.Component {
	r.initialize()
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return r.nodeRenderer.RenderNode(ctx, w, source, n)
	})
}

func (r *MarkdownRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	r.initialize()
	writer := getBufferedWriter(w)
	err := r.nodeRenderer.RenderNode(context.Background(), writer, source, n)
	if err != nil {
		return err
	}
	return writer.Flush()
}

func (r *MarkdownRenderer) initialize() {
	r.initSync.Do(func() {
		r.options = r.config.Options
		r.config.NodeRenderers.Sort()
		r.initializeNodeRenderers()
		r.config = nil
		r.nodeRendererFuncsTmp = nil
	})
}

func (r *MarkdownRenderer) initializeNodeRenderers() {
	l := len(r.config.NodeRenderers)
	for i := l - 1; i >= 0; i-- {
		v := r.config.NodeRenderers[i]
		nr, _ := v.Value.(renderer.NodeRenderer)
		if se, ok := v.Value.(renderer.SetOptioner); ok {
			for oname, ovalue := range r.options {
				se.SetOption(oname, ovalue)
			}
		}
		nr.RegisterFuncs(r)
	}
	r.nodeRendererFuncs = make([]renderer.NodeRendererFunc, r.maxKind+1)
	for kind, nr := range r.nodeRendererFuncsTmp {
		r.nodeRendererFuncs[kind] = nr
	}
}

// NodeRenderer handles different types of markdown nodes
type NodeRenderer struct {
	builder *ComponentBuilder
	layout  registry.Layout
	parent  *MarkdownRenderer // Reference to parent renderer for default rendering functions
}

func NewNodeRenderer(
	builder *ComponentBuilder,
	layout registry.Layout,
	parent *MarkdownRenderer,
) *NodeRenderer {
	return &NodeRenderer{
		builder: builder,
		layout:  layout,
		parent:  parent,
	}
}

// RenderNode is the main entry point for rendering nodes
func (nr *NodeRenderer) RenderNode(ctx context.Context, w io.Writer, source []byte, node ast.Node) error {
	writer := getBufferedWriter(w)

	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n.Kind() {
		case parser.KindMargoNode:
			return nr.renderMargoNode(ctx, writer, source, n, entering)
		case ast.KindHeading:
			return nr.renderHeading(ctx, writer, source, n, entering)
		case ast.KindParagraph:
			return nr.renderParagraph(ctx, writer, source, n, entering)
		case ast.KindThematicBreak:
			return nr.renderHr(ctx, writer, source, n, entering)
		case ast.KindLink:
			return nr.renderLink(ctx, writer, source, n, entering)
		case ast.KindList:
			return nr.renderList(ctx, writer, source, n, entering)
		case ast.KindListItem:
			return nr.renderListItem(ctx, writer, source, n, entering)
		default:
			return nr.renderDefault(writer, source, n, entering)
		}
	})

	if err != nil {
		return err
	}
	return writer.Flush()
}

// renderMargoNode handles Margo-specific node rendering
func (nr *NodeRenderer) renderMargoNode(ctx context.Context, w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*parser.Document)
	for _, child := range n.Children {
		if err := nr.renderComponent(ctx, w, child, nil); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkSkipChildren, nil
}

func (nr *NodeRenderer) renderList(ctx context.Context, w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	component, ok := nr.layout.Get("ul")
	if !ok {
		return nr.renderDefault(w, source, n, entering)
	}
	if !entering {
		return ast.WalkContinue, nil
	}

	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if err := nr.RenderNode(ctx, w, source, c); err != nil {
				return err
			}
		}
		return nil
	})

	ctx = templ.WithChildren(ctx, children)
	cmp, err := nr.builder.Build(component, n.Attributes())
	if err != nil {
		return ast.WalkStop, err
	}

	if err := cmp.Render(ctx, w); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

func (nr *NodeRenderer) renderListItem(ctx context.Context, w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.ListItem)
	component, ok := nr.layout.Get("li")
	if !ok {
		return nr.renderDefault(w, source, n, entering)
	}
	if !entering {
		return ast.WalkContinue, nil
	}

	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if err := nr.RenderNode(ctx, w, source, c); err != nil {
				return err
			}
		}
		return nil
	})

	ctx = templ.WithChildren(ctx, children)
	cmp, err := nr.builder.Build(component, n.Attributes())
	if err != nil {
		return ast.WalkStop, err
	}

	if err := cmp.Render(ctx, w); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

func (nr *NodeRenderer) renderLink(ctx context.Context, w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	component, ok := nr.layout.Get("a")
	if !ok {
		return nr.renderDefault(w, source, n, entering)
	}
	if !entering {
		return ast.WalkContinue, nil
	}

	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if err := nr.RenderNode(ctx, w, source, c); err != nil {
				return err
			}
		}
		return nil
	})

	ctx = templ.WithChildren(ctx, children)
	cmp, err := nr.builder.Build(component, append(n.Attributes(), ast.Attribute{
		Name:  []byte("Href"),
		Value: string(n.Destination),
	}))
	if err != nil {
		return ast.WalkStop, err
	}

	if err := cmp.Render(ctx, w); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

func (nr *NodeRenderer) renderHr(ctx context.Context, w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.ThematicBreak)
	component, ok := nr.layout.Get("hr")
	if !ok {
		return nr.renderDefault(w, source, n, entering)
	}
	if !entering {
		return ast.WalkContinue, nil
	}

	cmp, err := nr.builder.Build(component, nil)
	if err != nil {
		return ast.WalkStop, err
	}

	if err := cmp.Render(ctx, w); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

func (nr *NodeRenderer) renderParagraph(ctx context.Context, w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Paragraph)
	component, ok := nr.layout.Get("p")
	if !ok {
		return nr.renderDefault(w, source, n, entering)
	}
	if !entering {
		return ast.WalkContinue, nil
	}

	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if err := nr.RenderNode(ctx, w, source, c); err != nil {
				return err
			}
		}
		return nil
	})

	ctx = templ.WithChildren(ctx, children)
	cmp, err := nr.builder.Build(component, n.Attributes())
	if err != nil {
		return ast.WalkStop, err
	}

	if err := cmp.Render(ctx, w); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// renderHeading handles heading node rendering
func (nr *NodeRenderer) renderHeading(ctx context.Context, w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	component, ok := nr.layout.Get(fmt.Sprintf("h%d", n.Level))
	if !ok {
		return nr.renderDefault(w, source, n, entering)
	}

	if !entering {
		return ast.WalkContinue, nil
	}

	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if err := nr.RenderNode(ctx, w, source, child); err != nil {
				return err
			}
		}
		return nil
	})

	ctx = templ.WithChildren(ctx, children)
	cmp, err := nr.builder.Build(component, n.Attributes())
	if err != nil {
		return ast.WalkStop, err
	}

	if err := cmp.Render(ctx, w); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// renderDefault handles default node rendering
func (nr *NodeRenderer) renderDefault(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if f := nr.getDefaultRenderer(node.Kind()); f != nil {
		return f(w, source, node, entering)
	}
	return ast.WalkContinue, nil
}

// getDefaultRenderer returns the appropriate renderer function for a given node kind
func (nr *NodeRenderer) getDefaultRenderer(kind ast.NodeKind) renderer.NodeRendererFunc {
	// This method should be initialized with the default renderers from the parent MarkdownRenderer
	// We'll need to modify the MarkdownRenderer to expose its nodeRendererFuncs
	if kind >= 0 && int(kind) < len(nr.parent.nodeRendererFuncs) {
		return nr.parent.nodeRendererFuncs[kind]
	}
	return nil
}

// renderComponent handles component rendering
func (nr *NodeRenderer) renderComponent(ctx context.Context, w io.Writer, node parser.Node, parentNS registry.Layout) error {
	switch n := node.(type) {
	case *parser.ComponentNode:
		return nr.renderComponentNode(ctx, w, n, parentNS)
	case *parser.TextNode:
		return goldmark.Convert([]byte(n.Value), w)
	default:
		return fmt.Errorf("unsupported node type: %T", node)
	}
}

// renderComponentNode handles specific component node rendering
func (nr *NodeRenderer) renderComponentNode(
	ctx context.Context, w io.Writer,
	node *parser.ComponentNode,
	parentNS registry.Layout,
) error {
	if node.Name == "Slot" {
		return nr.renderSlot(ctx, w)
	}

	cmpFunc, err := nr.builder.GetComponent(node.Name, parentNS)
	if err != nil {
		return err
	}

	component, err := nr.builder.Build(cmpFunc, node.Attributes())
	if err != nil {
		return fmt.Errorf("failed to build component %s: %w", node.Name, err)
	}

	if parentNS == nil {
		parentNS = nr.layout
	}

	children, err := nr.renderChildren(node, parentNS)
	if err != nil {
		return fmt.Errorf("failed to render children for component %s: %w", node.Name, err)
	}

	return component.Render(templ.WithChildren(ctx, children), w)
}

// renderSlot handles slot rendering
func (nr *NodeRenderer) renderSlot(ctx context.Context, w io.Writer) error {
	component, ok := GetSlot(ctx)
	if !ok {
		return errors.New("slot not found")
	}
	return component.Render(ctx, w)
}

// renderChildren handles rendering of child nodes
func (nr *NodeRenderer) renderChildren(node *parser.ComponentNode, namespace registry.Layout) (templ.Component, error) {
	var components []templ.Component
	ns, err := namespace.Namespace(node.Name)
	if err != nil {
		ns, err = nr.layout.Namespace(node.Name)
	}
	for _, child := range node.Children() {
		switch c := child.(type) {
		case *parser.ComponentNode:
			comp := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				return nr.renderComponent(ctx, w, c, ns)
			})
			components = append(components, comp)
		case *parser.TextNode:
			comp := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
				return margo.New(nr.layout).Convert([]byte(c.Value), w)
			})
			components = append(components, comp)
		}
	}

	return templ.Join(components...), nil
}

// Helper function to get a buffered writer
func getBufferedWriter(w io.Writer) util.BufWriter {
	if bw, ok := w.(util.BufWriter); ok {
		return bw
	}
	return bufio.NewWriter(w)
}
