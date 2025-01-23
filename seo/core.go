package seo

import (
	"context"
	"fmt"
	"io"
)

var (
	MetaTags = map[string]string{
		"description":         "description",
		"author":              "author",
		"og_title":            "og:title",
		"og_description":      "og:description",
		"og_image":            "og:image",
		"og_url":              "og:url",
		"og_site_name":        "og:site_name",
		"og_type":             "og:type",
		"og_locale":           "og:locale",
		"twitter_card":        "twitter:card",
		"twitter_site":        "twitter:site",
		"twitter_title":       "twitter:title",
		"twitter_description": "twitter:description",
		"twitter_image":       "twitter:image",
	}
)

type Meta struct {
	Title string
	Tags  []*Tag
}

func (m *Meta) AddTag(name, value string) {
	m.Tags = append(m.Tags, &Tag{Name: name, Value: value})
}

func (m *Meta) Render(ctx context.Context, w io.Writer) error {
	for _, t := range m.Tags {
		if err := t.Render(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

type Tag struct {
	Name  string
	Value string
}

func (t *Tag) String() string {
	return fmt.Sprintf("<meta name=\"%s\" content=\"%s\"/>", t.Name, t.Value)
}

func (t *Tag) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(t.String()))
	return err
}

func FromPageMeta(meta map[string]any) *Meta {
	seo := &Meta{}

	if title, ok := meta["title"].(string); ok {
		seo.Title = title
	}

	for metaKey, tagName := range MetaTags {
		if value, ok := meta[metaKey].(string); ok {
			seo.AddTag(tagName, value)
		}
	}

	return seo
}
