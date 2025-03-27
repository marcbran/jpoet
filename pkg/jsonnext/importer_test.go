package jsonnext

import (
	"github.com/google/go-jsonnet"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFSImporter_ImportSuccess(t *testing.T) {
	fs := fstest.MapFS{
		"test.jsonnet": &fstest.MapFile{Data: []byte("local x = 1; x")},
	}
	importer := &FSImporter{Fs: fs}

	contents, foundAt, err := importer.Import("", "test.jsonnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if foundAt != "test.jsonnet" {
		t.Errorf("expected foundAt to be 'test.jsonnet', got %s", foundAt)
	}

	if !strings.Contains(contents.String(), "local x = 1") {
		t.Errorf("unexpected contents: %s", contents.String())
	}
}

func TestFSImporter_ImportNotExist(t *testing.T) {
	fs := fstest.MapFS{}
	importer := &FSImporter{Fs: fs}

	_, _, err := importer.Import("", "nonexistent.jsonnet")
	if err == nil || !strings.Contains(err.Error(), "couldn't open import") {
		t.Errorf("expected import error for nonexistent file, got: %v", err)
	}
}

func TestFSImporter_Cache(t *testing.T) {
	fs := fstest.MapFS{
		"cached.jsonnet": &fstest.MapFile{Data: []byte("local x = 1; x")},
	}
	importer := &FSImporter{Fs: fs}

	_, _, err := importer.Import("", "cached.jsonnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete the file from fs to simulate caching
	delete(fs, "cached.jsonnet")

	_, _, err = importer.Import("", "cached.jsonnet")
	if err != nil {
		t.Errorf("expected cached file, got error: %v", err)
	}
}

type mockImporter struct {
	content jsonnet.Contents
	path    string
	err     error
}

func (m mockImporter) Import(_, _ string) (jsonnet.Contents, string, error) {
	return m.content, m.path, m.err
}

func TestCompoundImporter_ImportSuccess(t *testing.T) {
	importer := CompoundImporter{
		Importers: []jsonnet.Importer{
			mockImporter{err: os.ErrNotExist},
			mockImporter{
				content: jsonnet.MakeContents("valid content"),
				path:    "valid.jsonnet",
				err:     nil,
			},
		},
	}

	contents, foundAt, err := importer.Import("", "test.jsonnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if foundAt != "valid.jsonnet" || contents.String() != "valid content" {
		t.Errorf("unexpected result: %s, %s", contents.String(), foundAt)
	}
}

func TestCompoundImporter_StopsOnSuccess(t *testing.T) {
	importer := CompoundImporter{
		Importers: []jsonnet.Importer{
			mockImporter{
				content: jsonnet.MakeContents("valid content"),
				path:    "valid.jsonnet",
				err:     nil,
			},
			mockImporter{err: os.ErrNotExist},
		},
	}

	contents, foundAt, err := importer.Import("", "test.jsonnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if foundAt != "valid.jsonnet" || contents.String() != "valid content" {
		t.Errorf("unexpected result: %s, %s", contents.String(), foundAt)
	}
}

func TestCompoundImporter_ContinuesOnNotExist(t *testing.T) {
	importer := CompoundImporter{
		Importers: []jsonnet.Importer{
			// First importer returns file not found
			mockImporter{err: os.ErrNotExist},

			// Second importer succeeds
			mockImporter{
				content: jsonnet.MakeContents("fallback content"),
				path:    "fallback.jsonnet",
				err:     nil,
			},
		},
	}

	contents, foundAt, err := importer.Import("", "test.jsonnet")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if contents.String() != "fallback content" || foundAt != "fallback.jsonnet" {
		t.Errorf("unexpected result: contents=%q, foundAt=%q", contents.String(), foundAt)
	}
}
