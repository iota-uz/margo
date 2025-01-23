package types

import (
	"context"
	"errors"
	"github.com/iota-uz/margo/seo"
	"net/http"
	"net/url"
)

type Key string

const PageCtxKey Key = "pageCtx"

var ErrNoPageCtx = errors.New("page context not found")

func NewPageCtx(r *http.Request, seo *seo.Meta) (*PageContext, error) {
	//locale := UseLocale(r.Context(), language.English)
	return &PageContext{
		URL:    r.URL,
		Locale: "en",
		Seo:    seo,
	}, nil
}

func WithPageCtx(ctx context.Context, pageCtx *PageContext) context.Context {
	return context.WithValue(ctx, PageCtxKey, pageCtx)
}

func UsePageCtx(ctx context.Context) (*PageContext, error) {
	pageCtx, ok := ctx.Value(PageCtxKey).(*PageContext)
	if !ok {
		return nil, ErrNoPageCtx
	}
	return pageCtx, nil
}

func MustUsePageCtx(ctx context.Context) *PageContext {
	pageCtx, err := UsePageCtx(ctx)
	if err != nil {
		panic(err)
	}
	return pageCtx
}

type PageContext struct {
	URL    *url.URL
	Locale string
	Seo    *seo.Meta
}
