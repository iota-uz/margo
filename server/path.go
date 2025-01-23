package server

import (
	"path/filepath"
	"strings"
)

// PathToUrl converts a staticFile path to a URL.
// Ex.: "foo/bar/index.html" -> "/foo/bar"
// Ex.: "foo/bar/blog.html" -> "/foo/bar/blog"
func PathToUrl(path string, ext string) string {
	dir, fn := filepath.Split(path)
	baseFn := StripExt(fn)
	if baseFn == "index" {
		return "/" + strings.TrimSuffix(dir, "/")
	}
	if filepath.Ext(path) == ext {
		return "/" + dir + baseFn
	}
	return "/" + path
}

// StripExt removes the extension from a file name.
// "index.html" -> "index"
// "bar.md" -> "bar"
func StripExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
