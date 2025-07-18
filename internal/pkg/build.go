package pkg

import (
	"context"
	"embed"
	"errors"
	"github.com/marcbran/gensonnet/pkg/gensonnet/config"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/formatter"
	"github.com/marcbran/devsonnet/internal/pkg/lib/imports"
	"github.com/marcbran/gensonnet/pkg/gensonnet"
)

func Build(ctx context.Context, pkgDir, buildDir string) error {
	mainFile := filepath.Join(pkgDir, "main.libsonnet")
	pkgFile := filepath.Join(pkgDir, "pkg.libsonnet")
	examplesFile := filepath.Join(pkgDir, "examples.libsonnet")

	_, err := os.Stat(mainFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return errors.New("main.libsonnet not found")
	}
	inlinedMainCode, err := inlineFile(mainFile)
	if err != nil {
		return err
	}

	_, err = os.Stat(pkgFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return errors.New("pkg.libsonnet not found")
	}
	pkgCode, err := os.ReadFile(pkgFile)
	if err != nil {
		return err
	}

	examplesCode := []byte("{}")
	_, err = os.Stat(examplesFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		examplesCode, err = os.ReadFile(examplesFile)
		if err != nil {
			return err
		}
	}

	err = gensonnet.RenderWithConfig(ctx, config.Config{
		Render: config.RenderConfig{
			TargetDir: buildDir,
			Lib: config.LibConfig{
				ManifestCode: `
					local lib = import 'input/lib.libsonnet';
					local libString = importstr 'input/lib.libsonnet';
					local pkg = import 'input/pkg.libsonnet';
					local examples = import 'input/examples.libsonnet';
					local examplesString = importstr 'input/examples.libsonnet';

					local manifest = import 'lib/manifest.libsonnet';
					manifest(lib, libString, pkg, examples, examplesString)
				`,
				Filesystems: []embed.FS{
					lib,
					imports.Fs,
				},
				Imports: map[string]string{
					"input/lib.libsonnet":      inlinedMainCode,
					"input/pkg.libsonnet":      string(pkgCode),
					"input/examples.libsonnet": string(examplesCode),
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func inlineFile(input string) (string, error) {
	node, finalFodder, err := readInlineNode(input)
	if err != nil {
		return "", err
	}
	format, err := formatter.FormatNode(node, finalFodder, formatter.DefaultOptions())
	if err != nil {
		return "", err
	}
	return format, nil
}

func readInlineNode(filename string) (ast.Node, ast.Fodder, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	node, finalFodder, err := formatter.SnippetToRawAST(filename, string(b))
	if err != nil {
		return nil, nil, err
	}
	err = inlineNode(node, filename)
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
			impNode, _, err := readInlineNode(impFilename)
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
