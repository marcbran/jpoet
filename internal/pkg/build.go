package pkg

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/marcbran/gensonnet/pkg/gensonnet/config"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/formatter"
	"github.com/marcbran/devsonnet/internal/pkg/lib/imports"
	"github.com/marcbran/gensonnet/pkg/gensonnet"
)

//go:embed lib
var lib embed.FS

func Build(ctx context.Context, pkgDir, outDir string) error {
	mainFile := filepath.Join(pkgDir, "main.libsonnet")
	pkgFile := filepath.Join(pkgDir, "pkg.libsonnet")

	outMainFile := filepath.Join(outDir, "main.libsonnet")

	_, err := os.Stat(mainFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return errors.New("main.libsonnet not found")
	}

	_, err = os.Stat(pkgFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return errors.New("pkg.libsonnet not found")
	}

	examplesImport := fmt.Sprintf("import '%s/examples.libsonnet'", pkgDir)
	examplesStringImport := fmt.Sprintf("importstr '%s/examples.libsonnet'", pkgDir)
	examplesFile := filepath.Join(pkgDir, "examples.libsonnet")
	_, err = os.Stat(examplesFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		examplesImport = "null"
		examplesStringImport = "null"
	}

	err = os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = inlineFile(mainFile, outMainFile)
	if err != nil {
		return err
	}

	err = gensonnet.RenderWithConfig(ctx, config.Config{
		Render: config.RenderConfig{
			TargetDir: outDir,
			Lib: config.LibConfig{
				ManifestStr: fmt.Sprintf(`
				local lib = import '%s/main.libsonnet';
				local pkg = import '%s/pkg.libsonnet';
				local examples = %s;
				local examplesString = %s;

				local manifest = import 'lib/main.libsonnet';
				manifest(lib, pkg, examples, examplesString)
				`, pkgDir, pkgDir, examplesImport, examplesStringImport),
				Filesystems: []embed.FS{
					lib,
					imports.Fs,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func inlineFile(input string, output string) error {
	node, finalFodder, err := readRawAST(input)
	if err != nil {
		return err
	}

	err = inlineNode(node, input)
	if err != nil {
		return err
	}

	err = writeFormattedNode(node, finalFodder, output)
	if err != nil {
		return err
	}
	return nil
}

func readRawAST(filename string) (ast.Node, ast.Fodder, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	node, finalFodder, err := formatter.SnippetToRawAST(filename, string(b))
	if err != nil {
		return nil, nil, err
	}
	return node, finalFodder, nil
}

func inlineNode(node ast.Node, filename string) error {
	if local, ok := node.(*ast.Local); ok {
		if imp, ok := local.Binds[0].Body.(*ast.Import); ok {
			impFilename, err := filepath.Abs(filepath.Join(filepath.Dir(filename), imp.File.Value))
			if err != nil {
				return err
			}
			impNode, _, err := readRawAST(impFilename)
			if err != nil {
				return err
			}
			err = inlineNode(impNode, impFilename)
			if err != nil {
				return err
			}
			if nodeBase, ok := impNode.(*ast.Local); ok {
				nodeBase.Fodder = []ast.FodderElement{
					{Kind: ast.FodderLineEnd},
				}
			}
			local.Binds[0].Body = impNode
		}
		return inlineNode(local.Body, filename)
	}
	return nil
}

func writeFormattedNode(node ast.Node, finalFodder ast.Fodder, filename string) error {
	format, err := formatter.FormatNode(node, finalFodder, formatter.DefaultOptions())
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, []byte(format), 0644)
	if err != nil {
		return err
	}
	return nil
}
