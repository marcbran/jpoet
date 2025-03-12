package test

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/marcbran/jsonnet-kit/internal/jsonnext"
	"io/fs"
	"path/filepath"
	"strings"
)

type Run struct {
	Results     []Result `json:"results"`
	PassedCount int      `json:"passedCount"`
	TotalCount  int      `json:"totalCount"`
}

func (r Run) append(o Run) Run {
	return Run{
		Results:     append(r.Results, o.Results...),
		PassedCount: r.PassedCount + o.PassedCount,
		TotalCount:  r.TotalCount + o.TotalCount,
	}
}

type Result struct {
	Name     string `json:"name"`
	Expected any    `json:"expected"`
	Actual   any    `json:"actual"`
	Equal    bool   `json:"equal"`
}

//go:embed lib
var lib embed.FS

func RunDir(dirname string) (*Run, error) {
	var run Run
	var runErr error
	err := filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !strings.HasSuffix(path, "_tests.libsonnet") {
			return nil
		}
		r, err := RunFile(path)
		if err != nil {
			runErr = err
			fmt.Println()
			fmt.Println(err.Error())
			return nil
		}
		if r != nil {
			run = run.append(*r)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if runErr != nil {
		return nil, errors.New("encountered at least one error while running tests")
	}
	return &run, nil
}

func RunFile(filename string) (*Run, error) {
	fmt.Printf("  File: %s\n", filename)

	vm := jsonnet.MakeVM()
	vm.Importer(jsonnext.CompoundImporter{
		Importers: []jsonnet.Importer{
			&jsonnext.FSImporter{Fs: lib},
			&jsonnet.FileImporter{},
		},
	})
	res, err := vm.EvaluateAnonymousSnippet("main.jsonnet", fmt.Sprintf(`
		local tests = import '%s';
		local lib = import 'lib/main.libsonnet';
		lib.runTests(tests)
	`, filename))
	if err != nil {
		return nil, err
	}
	var run Run
	err = json.Unmarshal([]byte(res), &run)
	if err != nil {
		return nil, err
	}
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
	return &run, nil
}
