package ssg

import (
	"fmt"
	"github.com/iota-uz/margo/registry"
	"github.com/iota-uz/margo/server"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WatchOptions struct {
	SourceDir      string
	DestinationDir string
	Registry       registry.Registry
}

func countPages(items []*server.FsItem) int {
	count := 0
	for _, mdItem := range items {
		if !mdItem.IsStatic {
			count++
		}
	}
	return count
}

func Generate(src, dest string, reg registry.Registry) error {
	start := time.Now()
	items, err := server.IndexDirectory(os.DirFS(src), ".")
	if err != nil {
		return fmt.Errorf("failed to load items: %w", err)
	}
	if err := newGenerator(src, dest, reg).Generate(dest, items); err != nil {
		return err
	}
	log.Printf("Generated %d pages in %v\n", countPages(items), time.Since(start))
	return nil
}

func Watch(opts WatchOptions) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Println(err)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = filepath.WalkDir(opts.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		return watcher.Add(filepath.Join(wd, path))
	})
	if err != nil {
		return err
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}
			if event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				log.Println("Regenerating...")
				if err := Generate(opts.SourceDir, opts.DestinationDir, opts.Registry); err != nil {
					log.Println(err)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Println("error:", err)
		}
	}
}
