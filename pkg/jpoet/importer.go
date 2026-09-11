package jpoet

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/go-jsonnet"
)

type CompoundImporter struct {
	Importers []jsonnet.Importer
}

func (c CompoundImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	var errs []error
	for _, importer := range c.Importers {
		contents, foundAt, err = importer.Import(importedFrom, importedPath)
		if err == nil {
			return contents, foundAt, nil
		}
		errs = append(errs, err)
	}
	return contents, foundAt, errors.Join(errs...)
}

type MemoryImporter struct {
	Data map[string]jsonnet.Contents
}

func (importer *MemoryImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	dir, _ := filepath.Split(importedFrom)
	absPath := filepath.Join(dir, importedPath)
	if content, ok := importer.Data[absPath]; ok {
		return content, importedPath, nil
	}
	if content, ok := importer.Data[importedPath]; ok {
		return content, importedPath, nil
	}
	return jsonnet.Contents{}, "", fmt.Errorf("import not available %v", importedPath)
}

type FSImporter struct {
	Fs fs.FS

	cache importCache
}

func (importer *FSImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	dir, _ := filepath.Split(importedFrom)
	absPath := filepath.Join(dir, importedPath)
	contents, foundAt, err := importer.tryPath(absPath)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}
	if foundAt != "" {
		return contents, foundAt, nil
	}

	contents, foundAt, err = importer.tryPath(importedPath)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}
	if foundAt != "" {
		return contents, foundAt, nil
	}
	return jsonnet.Contents{}, "", fmt.Errorf("couldn't open import %#v: no match in provided file system", importedPath)
}

func (importer *FSImporter) tryPath(p string) (jsonnet.Contents, string, error) {
	if contents, exists, cached := importer.cache.get(p); cached {
		if !exists {
			return jsonnet.Contents{}, "", nil
		}
		return contents, p, nil
	}

	contentBytes, err := fs.ReadFile(importer.Fs, p)
	if err != nil {
		if !os.IsNotExist(err) {
			return jsonnet.Contents{}, "", err
		}
		importer.cache.set(p, jsonnet.Contents{}, false)
		return jsonnet.Contents{}, "", nil
	}

	contents := jsonnet.MakeContentsRaw(contentBytes)
	importer.cache.set(p, contents, true)
	return contents, p, nil
}

func (importer *FSImporter) FlushCache() {
	importer.cache.flush()
}

type FileImporter struct {
	JPaths []string

	cache importCache
}

func (importer *FileImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	dir, _ := filepath.Split(importedFrom)
	found, contents, foundAt, err := importer.tryPath(dir, importedPath)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}
	for i := len(importer.JPaths) - 1; !found && i >= 0; i-- {
		found, contents, foundAt, err = importer.tryPath(importer.JPaths[i], importedPath)
		if err != nil {
			return jsonnet.Contents{}, "", err
		}
	}
	if !found {
		return jsonnet.Contents{}, "", fmt.Errorf("couldn't open import %#v: no match locally or in the Jsonnet library paths", importedPath)
	}
	return contents, foundAt, nil
}

func (importer *FileImporter) tryPath(dir, importedPath string) (bool, jsonnet.Contents, string, error) {
	var absPath string
	if filepath.IsAbs(importedPath) {
		absPath = importedPath
	} else {
		absPath = filepath.Join(dir, importedPath)
	}

	if contents, exists, cached := importer.cache.get(absPath); cached {
		if !exists {
			return false, jsonnet.Contents{}, "", nil
		}
		return true, contents, absPath, nil
	}

	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, jsonnet.Contents{}, "", err
		}
		importer.cache.set(absPath, jsonnet.Contents{}, false)
		return false, jsonnet.Contents{}, "", nil
	}

	contents := jsonnet.MakeContentsRaw(contentBytes)
	importer.cache.set(absPath, contents, true)
	return true, contents, absPath, nil
}

func (importer *FileImporter) FlushCache() {
	importer.cache.flush()
}

type importCacheEntry struct {
	contents jsonnet.Contents
	exists   bool
}

type importCache struct {
	mu      sync.Mutex
	entries map[string]*importCacheEntry
}

func (c *importCache) get(path string) (jsonnet.Contents, bool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, cached := c.entries[path]
	if !cached {
		return jsonnet.Contents{}, false, false
	}
	return entry.contents, entry.exists, true
}

func (c *importCache) set(path string, contents jsonnet.Contents, exists bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = make(map[string]*importCacheEntry)
	}
	c.entries[path] = &importCacheEntry{contents: contents, exists: exists}
}

func (c *importCache) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = nil
}
