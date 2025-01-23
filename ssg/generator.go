package ssg

import (
	"context"
	"fmt"
	"github.com/iota-uz/margo/seo"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/a-h/templ"

	"github.com/iota-uz/margo/layouts"
	"github.com/iota-uz/margo/registry"
	"github.com/iota-uz/margo/server"
	"github.com/iota-uz/margo/types"
)

// GenerationError type to handle generation errors
type GenerationError struct {
	Path string
	Err  error
}

func (e *GenerationError) Error() string {
	return fmt.Sprintf("error generating %s: %v", e.Path, e.Err)
}

func newGenerator(src, dest string, reg registry.Registry) *generator {
	return &generator{
		loader:   server.NewLoader(os.DirFS(src)),
		src:      src,
		dest:     dest,
		registry: reg,
	}
}

type generator struct {
	registry registry.Registry
	loader   *server.MarkdownLoader
	src      string
	dest     string
}

func (g *generator) RenderPage(ctx context.Context, page server.Page) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("could not render page %s: %v", page.Path(), r)
		}
	}()
	u, err := url.Parse(page.URL())
	if err != nil {
		return "", err
	}
	ctx = types.WithPageCtx(ctx, &types.PageContext{
		URL:    u,
		Locale: "en",
		Seo:    seo.FromPageMeta(page.Meta()),
	})
	ctx = templ.WithChildren(ctx, page)
	var b strings.Builder
	if err := layouts.Base().Render(ctx, &b); err != nil {
		return "", err
	}
	return b.String(), nil
}

// processItem handles a single item's generation
func (g *generator) processItem(
	ctx context.Context,
	item *server.FsItem,
	dest string,
	errCh chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	if item.IsStatic {
		destFile := filepath.Join(dest, item.Path)
		if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
			errCh <- &GenerationError{Path: destFile, Err: err}
			return
		}
		if err := CopyFile(filepath.Join(g.src, item.Path), destFile); err != nil {
			errCh <- &GenerationError{Path: item.Path, Err: fmt.Errorf("failed to copy file: %w", err)}
			return
		}
	} else {
		page, err := g.loader.Load(item, g.registry)
		if err != nil {
			errCh <- &GenerationError{Path: item.Path, Err: err}
			return
		}
		content, err := g.RenderPage(ctx, page)
		if err != nil {
			errCh <- &GenerationError{Path: item.Path, Err: err}
			return
		}
		destFile := filepath.Join(dest, filepath.Dir(page.Path()), page.Name()+".html")
		if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
			errCh <- &GenerationError{Path: destFile, Err: err}
			return
		}
		if err := os.WriteFile(destFile, []byte(content), os.ModePerm); err != nil {
			errCh <- &GenerationError{Path: destFile, Err: err}
			return
		}
	}
}

// Generate processes all items concurrently
func (g *generator) Generate(dest string, items []*server.FsItem) error {
	if len(items) == 0 {
		return nil
	}

	// Create error channel and WaitGroup
	errCh := make(chan error, len(items))
	var wg sync.WaitGroup

	// Process items concurrently
	for _, item := range items {
		wg.Add(1)
		go g.processItem(context.Background(), item, dest, errCh, &wg)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Collect any errors
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		// Combine all errors into a single error message
		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf("generation errors:\n%s", strings.Join(errMsgs, "\n"))
	}

	return nil
}
