package server

import (
	"context"
	"github.com/a-h/templ"
	"io"
)

type Page interface {
	templ.Component
	// Name returns the name of the page. Ex.: "introduction"
	Name() string

	// Path returns the relative path of the page.
	// Ex.: "docs/introduction.md"
	Path() string

	// URL returns the page url mapped from the filesystem.
	// Ex.: "docs/introduction.md" -> "/docs/introduction"
	// Ex.: "docs/index.md" -> "/docs"
	URL() string

	// Meta returns the metadata of the page.
	// ---
	// Title: goldmark-meta
	// Summary: Add YAML metadata to the document
	// ---
	// tuns into
	// map[string]any{
	// 	"Title": "goldmark-meta",
	// 	"Summary": "Add YAML metadata to the document",
	// }
	Meta() map[string]any
}

var _ Page = &page{}

type page struct {
	name      string
	path      string
	url       string
	component templ.Component
	meta      map[string]any
}

func (f *page) Path() string {
	return f.path
}

func (f *page) Name() string {
	return f.name
}

func (f *page) Render(ctx context.Context, w io.Writer) error {
	return f.component.Render(ctx, w)
}

func (f *page) URL() string {
	return f.url
}

func (f *page) Meta() map[string]any {
	return f.meta
}
