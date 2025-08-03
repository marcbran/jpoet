package pkg

import (
	"context"
	"errors"
	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-jsonnet/jsonnet"
	"github.com/marcbran/jsonnet-plugin-markdown/markdown"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/formatter"
	"github.com/marcbran/jpoet/internal/pkg/lib/imports"
)

func Build(ctx context.Context, pkgDir, buildDir string) error {
	mainFile := filepath.Join(pkgDir, "main.libsonnet")
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

	pkgFile := filepath.Join(pkgDir, "pkg.libsonnet")
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

	examplesFile := filepath.Join(pkgDir, "examples.libsonnet")
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

	err = jpoet.NewEval().
		FileImport([]string{}).
		FSImport(lib).
		FSImport(imports.Fs).
		StringImport("input/main.libsonnet", inlinedMainCode).
		StringImport("input/pkg.libsonnet", string(pkgCode)).
		StringImport("input/examples.libsonnet", string(examplesCode)).
		Plugin(markdown.Plugin()).
		Plugin(jsonnet.Plugin()).
		TLACode("lib", "import 'input/main.libsonnet'").
		TLACode("libString", "importstr 'input/main.libsonnet'").
		TLACode("pkg", "import 'input/pkg.libsonnet'").
		TLACode("examples", "import 'input/examples.libsonnet'").
		TLACode("examplesString", "importstr 'input/examples.libsonnet'").
		FileInput("./lib/manifest.libsonnet").
		Serialize(false).
		DirectoryOutput(buildDir).
		Eval()
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
