package test

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/pkg/jpoet"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Run struct {
	Results     []Result `json:"results"`
	PassedCount int      `json:"passedCount"`
	TotalCount  int      `json:"totalCount"`
}

func (r Run) append(prefix string, o Run) Run {
	var other []Result
	for _, result := range o.Results {
		other = append(other, Result{
			Name:     fmt.Sprintf("%s%s", prefix, result.Name),
			Expected: result.Expected,
			Actual:   result.Actual,
			Equal:    result.Equal,
		})
	}
	return Run{
		Results:     append(r.Results, other...),
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

type Option func(*config)

type config struct {
	plugins []*jpoet.Plugin
}

// WithPlugins registers plugins so tests can call them through std.native.
func WithPlugins(plugins ...*jpoet.Plugin) Option {
	return func(c *config) {
		c.plugins = append(c.plugins, plugins...)
	}
}

func RunDir(dirname string, opts ...Option) (*Run, error) {
	var run Run
	var runErr error
	err := filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !strings.HasSuffix(path, "_tests.libsonnet") {
			return nil
		}
		r, err := RunFile(path, opts...)
		if err != nil {
			runErr = err
			_, err := os.Stderr.WriteString(err.Error())
			if err != nil {
				return err
			}
			return nil
		}
		if r != nil {
			run = run.append(path, *r)
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

func RunFile(filename string, opts ...Option) (*Run, error) {
	c := config{}
	for _, opt := range opts {
		opt(&c)
	}
	vm := jsonnet.MakeVM()
	for _, p := range c.plugins {
		vm.NativeFunction(p.NativeFunction())
	}
	vm.Importer(jpoet.CompoundImporter{
		Importers: []jsonnet.Importer{
			&jpoet.FSImporter{Fs: lib},
			&jpoet.FileImporter{},
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
	return &run, nil
}
