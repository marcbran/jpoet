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

// Import TODO this is not testing well
func (importer *FSImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	if importer.fsCache == nil {
		importer.fsCache = make(map[string]*fsCacheEntry)
	}

	dir, _ := filepath.Split(importedFrom)
	absPath := filepath.Join(dir, importedPath)
	if cacheEntry, isCached := importer.fsCache[absPath]; isCached {
		if !cacheEntry.exists {
			return jsonnet.Contents{}, "", fmt.Errorf("couldn't open import %#v: no match in provided file system", importedPath)
		}
		return cacheEntry.contents, absPath, nil
	}

	contentBytes, err := fs.ReadFile(importer.Fs, absPath)

	if err != nil {
		if !os.IsNotExist(err) {
			return jsonnet.Contents{}, "", err
		}

		entry := &fsCacheEntry{
			exists: false,
		}
		importer.fsCache[absPath] = entry
		return jsonnet.Contents{}, "", fmt.Errorf("couldn't open import %#v: no match in provided file system", importedPath)
	}

	entry := &fsCacheEntry{
		exists:   true,
		contents: jsonnet.MakeContentsRaw(contentBytes),
	}
	importer.fsCache[absPath] = entry
	return entry.contents, absPath, nil
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
