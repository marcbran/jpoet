package pkg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-jsonnet/jsonnet"
	"github.com/marcbran/jsonnet-plugin-markdown/markdown"

	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/formatter"
	"github.com/marcbran/jpoet/internal/pkg/lib/imports"
)

func Build(ctx context.Context, pkgDir, buildDir string, jpaths []string) error {
	mainFile := filepath.Join(pkgDir, "main.libsonnet")
	_, err := os.Stat(mainFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return errors.New("main.libsonnet not found")
	}

	cfg, err := ResolvePkgConfig(ctx, pkgDir)
	if err != nil {
		return err
	}
	in := inliner{external: make(map[string]bool, len(cfg.External)), jpaths: jpaths}
	for _, imp := range cfg.External {
		in.external[imp] = true
	}

	inlinedMainCode, err := in.inlineFile(mainFile)
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

	err = jpoet.Eval(ctx,
		jpoet.FileImport([]string{}),
		jpoet.FSImport(lib),
		jpoet.FSImport(imports.Fs),
		jpoet.StringImport("input/main.libsonnet", inlinedMainCode),
		jpoet.StringImport("input/pkg.libsonnet", string(pkgCode)),
		jpoet.StringImport("input/examples.libsonnet", string(examplesCode)),
		jpoet.WithPlugin(markdown.Plugin()),
		jpoet.WithPlugin(jsonnet.Plugin()),
		jpoet.TLACode("lib", "import 'input/main.libsonnet'"),
		jpoet.TLACode("libString", "importstr 'input/main.libsonnet'"),
		jpoet.TLACode("pkg", "import 'input/pkg.libsonnet'"),
		jpoet.TLACode("examples", "import 'input/examples.libsonnet'"),
		jpoet.TLACode("examplesString", "importstr 'input/examples.libsonnet'"),
		jpoet.FileInput("./lib/manifest.libsonnet"),
		jpoet.Serialize(false),
		jpoet.DirectoryOutput(buildDir),
	)
	if err != nil {
		return err
	}
	return nil
}

type inliner struct {
	external map[string]bool
	jpaths   []string
}

func (in inliner) inlineFile(input string) (string, error) {
	node, finalFodder, err := in.readInlineNode(input)
	if err != nil {
		return "", err
	}
	format, err := formatter.FormatNode(node, finalFodder, formatter.DefaultOptions())
	if err != nil {
		return "", err
	}
	return format, nil
}

func (in inliner) readInlineNode(filename string) (ast.Node, ast.Fodder, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	node, finalFodder, err := formatter.SnippetToRawAST(filename, string(b))
	if err != nil {
		return nil, nil, err
	}
	err = in.inlineNode(node, filename)
	if err != nil {
		return nil, nil, err
	}
	return node, finalFodder, nil
}

func (in inliner) inlineNode(node ast.Node, filename string) error {
	switch n := node.(type) {
	case *ast.Local:
		for i := range n.Binds {
			resolved, err := in.inlineValue(n.Binds[i].Body, filename)
			if err != nil {
				return err
			}
			n.Binds[i].Body = resolved
		}
		return in.inlineNode(n.Body, filename)
	case *ast.Object:
		for i := range n.Fields {
			if n.Fields[i].Expr2 == nil {
				continue
			}
			resolved, err := in.inlineValue(n.Fields[i].Expr2, filename)
			if err != nil {
				return err
			}
			n.Fields[i].Expr2 = resolved
		}
	}
	return nil
}

func (in inliner) inlineValue(node ast.Node, filename string) (ast.Node, error) {
	switch imp := node.(type) {
	case *ast.Import:
		return in.inlineImport(imp, filename)
	case *ast.ImportStr:
		return in.inlineImportStr(imp, filename)
	default:
		return node, nil
	}
}

func (in inliner) inlineImport(imp *ast.Import, filename string) (ast.Node, error) {
	if in.external[imp.File.Value] {
		return imp, nil
	}
	impFilename, err := in.importFilename(imp.File.Value, filename)
	if err != nil {
		return nil, err
	}
	impNode, _, err := in.readInlineNode(impFilename)
	if err != nil {
		return nil, err
	}
	if nodeBase, ok := impNode.(*ast.Local); ok {
		nodeBase.Fodder = []ast.FodderElement{
			{Kind: ast.FodderLineEnd},
		}
	}
	return impNode, nil
}

func (in inliner) inlineImportStr(imp *ast.ImportStr, filename string) (ast.Node, error) {
	if in.external[imp.File.Value] {
		return imp, nil
	}
	impFilename, err := in.importFilename(imp.File.Value, filename)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(impFilename)
	if err != nil {
		return nil, err
	}
	value := string(b)
	if !strings.HasSuffix(value, "\n") {
		value += "\n"
	}
	lit := &ast.LiteralString{
		Value: value,
		Kind:  ast.StringBlock,
	}
	*lit.OpenFodder() = *imp.OpenFodder()
	return lit, nil
}

func (in inliner) importFilename(file, filename string) (string, error) {
	local, err := filepath.Abs(filepath.Join(filepath.Dir(filename), file))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}
	for i := len(in.jpaths) - 1; i >= 0; i-- {
		candidate, err := filepath.Abs(filepath.Join(in.jpaths[i], file))
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return local, nil
}
