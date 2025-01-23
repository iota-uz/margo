package server

import "testing"

func TestPathToUrl(t *testing.T) {
	tests := []struct {
		path string
		url  string
	}{
		{
			path: "index.html",
			url:  "/",
		},
		{
			path: "foo/index.html",
			url:  "/foo",
		},
		{
			path: "foo/bar/blog.html",
			url:  "/foo/bar/blog",
		},
		{
			path: "foo/bar/index.html",
			url:  "/foo/bar",
		},
		{
			path: "foo/bar/baz.css",
			url:  "/foo/bar/baz.css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := PathToUrl(tt.path, ".html"); got != tt.url {
				t.Errorf("expected: %s, got: %s", tt.url, got)
			}
		})
	}
}
