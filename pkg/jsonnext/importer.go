package jsonnext

import (
	"fmt"
	"github.com/google/go-jsonnet"
	"io/fs"
	"os"
	"path/filepath"
)

type FSImporter struct {
	Fs      fs.FS
	fsCache map[string]*fsCacheEntry
}

type fsCacheEntry struct {
	contents jsonnet.Contents
	exists   bool
}

func (importer *FSImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	if importer.fsCache == nil {
		importer.fsCache = make(map[string]*fsCacheEntry)
	}

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
	if cacheEntry, isCached := importer.fsCache[p]; isCached {
		if !cacheEntry.exists {
			return jsonnet.Contents{}, "", nil
		}
		return cacheEntry.contents, p, nil
	}

	contentBytes, err := fs.ReadFile(importer.Fs, p)

	if err != nil {
		if !os.IsNotExist(err) {
			return jsonnet.Contents{}, "", err
		}

		importer.fsCache[p] = &fsCacheEntry{
			exists: false,
		}
		return jsonnet.Contents{}, "", nil
	}

	entry := &fsCacheEntry{
		exists:   true,
		contents: jsonnet.MakeContentsRaw(contentBytes),
	}
	importer.fsCache[p] = entry
	return entry.contents, p, nil
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

type CompoundImporter struct {
	Importers []jsonnet.Importer
}

func (c CompoundImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	for _, importer := range c.Importers {
		contents, foundAt, err = importer.Import(importedFrom, importedPath)
		if err == nil {
			break
		}
	}
	return contents, foundAt, err
}
