package internal

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/google/go-jsonnet"
	"io/fs"
	"path/filepath"
	"strings"
)

type TestRun struct {
	Results     []TestResult `json:"results"`
	PassedCount int          `json:"passedCount"`
	TotalCount  int          `json:"totalCount"`
}

type TestResult struct {
	Name     string `json:"name"`
	Expected any    `json:"expected"`
	Actual   any    `json:"actual"`
	Equal    bool   `json:"equal"`
}

//go:embed lib
var lib embed.FS

func TestDir(dirname string) error {
	return filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !strings.HasSuffix(path, "_tests.libsonnet") {
			return nil
		}
		err = TestFile(path)
		if err != nil {
			return err
		}
		return nil
	})
}

func TestFile(filename string) error {
	vm := jsonnet.MakeVM()
	vm.Importer(CompoundImporter{
		Importers: []jsonnet.Importer{
			&FSImporter{fs: lib},
			&jsonnet.FileImporter{},
		},
	})
	res, err := vm.EvaluateAnonymousSnippet("main.jsonnet", fmt.Sprintf(`
		local tests = import '%s';
		local lib = import 'lib/main.libsonnet';
		lib.runTests(tests)
	`, filename))
	if err != nil {
		return err
	}
	var run TestRun
	err = json.Unmarshal([]byte(res), &run)
	if err != nil {
		return err
	}
	fmt.Printf("  File: %s\n", filename)
	if run.PassedCount < run.TotalCount {
		fmt.Println()
	}
	for _, result := range run.Results {
		if !result.Equal {
			fmt.Printf("      Name: %s\n", result.Name)
			fmt.Printf("    Actual: %s\n", result.Actual)
			fmt.Printf("  Expected: %s\n", result.Expected)
			fmt.Println()
		}
	}
	fmt.Printf("Passed: %d/%d\n", run.PassedCount, run.TotalCount)
	fmt.Println()
	return nil
}
