package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/iota-uz/margo/renderer"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"

	"github.com/iota-uz/margo"
	"github.com/iota-uz/margo/registry"
)

// ErrLayoutNotFound is returned when a specified layout cannot be found in the registry
type ErrLayoutNotFound struct {
	Layout           string
	AvailableLayouts []string
}

func (e ErrLayoutNotFound) Error() string {
	return fmt.Sprintf("layout %q not found. Must be one of the following: %s",
		e.Layout, strings.Join(e.AvailableLayouts, ", "))
}

// ErrInvalidLayoutType is returned when the layout metadata is not a string
var ErrInvalidLayoutType = errors.New("layout must be a string")

// ErrMissingLayout is returned when no layout is specified in the metadata
type ErrMissingLayout struct {
	AvailableLayouts []string
}

func (e ErrMissingLayout) Error() string {
	return fmt.Sprintf("missing layout in page meta. Must be one of the following: %s",
		strings.Join(e.AvailableLayouts, ", "))
}

var (
	LayoutFile = "layout.md"
)

func NewLoader(fsys fs.FS) *MarkdownLoader {
	return &MarkdownLoader{
		fs: fsys,
	}
}

func GetMeta(content []byte) (map[string]any, error) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)
	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := markdown.Convert(content, &buf, parser.WithContext(ctx)); err != nil {
		return nil, err
	}
	return meta.Get(ctx), nil
}

type MarkdownLoader struct {
	fs fs.FS
}

func (m *MarkdownLoader) GetMeta(path string) (map[string]any, error) {
	fileBytes, err := fs.ReadFile(m.fs, path)
	if err != nil {
		return nil, err
	}
	return GetMeta(fileBytes)
}

func (m *MarkdownLoader) useLayout(reg registry.Registry, fileMeta map[string]any) (registry.Layout, error) {
	layoutMeta, exists := fileMeta["layout"]
	if !exists {
		return nil, &ErrMissingLayout{
			AvailableLayouts: reg.Layouts(),
		}
	}

	layoutName, ok := layoutMeta.(string)
	if !ok {
		return nil, ErrInvalidLayoutType
	}

	layout, found := reg.Use(layoutName)
	if !found {
		return nil, &ErrLayoutNotFound{
			Layout:           layoutName,
			AvailableLayouts: reg.Layouts(),
		}
	}

	return layout, nil
}

func (m *MarkdownLoader) Load(item *FsItem, reg registry.Registry) (Page, error) {
	fileBytes, err := fs.ReadFile(m.fs, item.Path)
	if err != nil {
		return nil, err
	}
	fileMeta, err := GetMeta(fileBytes)
	if err != nil {
		return nil, err
	}
	layout, err := m.useLayout(reg, fileMeta)
	if err != nil {
		return nil, err
	}
	margoConverter := margo.New(layout)
	var component templ.Component
	if item.Layout == "" {
		component = margoConverter.ConvertToTempl(fileBytes)
	} else {
		layoutBytes, err := fs.ReadFile(m.fs, item.Layout)
		if err != nil {
			return nil, err
		}
		component = templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			return margoConverter.ConvertToTempl(layoutBytes).Render(
				renderer.WithSlot(ctx, margoConverter.ConvertToTempl(fileBytes)),
				w,
			)
		})
	}
	name := StripExt(filepath.Base(item.Path))
	return &page{
		name:      name,
		path:      item.Path,
		url:       item.URL,
		component: component,
		meta:      fileMeta,
	}, nil
}

type FsItem struct {
	IsStatic bool
	Path     string
	Layout   string
	URL      string
}

func IndexDirectory(fsys fs.FS, dir string) ([]*FsItem, error) {
	var result []*FsItem
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	var layout string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == LayoutFile {
			layout = filepath.Join(dir, entry.Name())
		}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(dir, entry.Name())
			children, err := IndexDirectory(fsys, fullPath)
			if err != nil {
				return nil, err
			}
			result = append(result, children...)
		} else {
			if entry.Name() == LayoutFile {
				continue
			}
			extension := filepath.Ext(entry.Name())
			path := filepath.Join(dir, entry.Name())
			if strings.ToLower(extension) != ".md" {
				result = append(result, &FsItem{
					Path:     path,
					Layout:   "",
					IsStatic: true,
					URL:      PathToUrl(path, ".md"),
				})
			} else {
				result = append(result, &FsItem{
					Path:     path,
					Layout:   layout,
					IsStatic: false,
					URL:      PathToUrl(path, ".md"),
				})
			}
		}
	}
	return result, nil
}
