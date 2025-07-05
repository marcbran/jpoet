package bundle

import (
	"context"
	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/formatter"
	"os"
	"path/filepath"
)

func Run(ctx context.Context, input, output string) error {
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
