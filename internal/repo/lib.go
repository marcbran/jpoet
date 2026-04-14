package repo

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/marcbran/jpoet/internal/pkg/lib/imports"
	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-jsonnet/jsonnet"
	"github.com/marcbran/jsonnet-plugin-markdown/markdown"
	"os"
)

//go:embed lib
var lib embed.FS

func manifestRepo(ctx context.Context, files map[string]string) (string, error) {
	b, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	buildDir, err := os.MkdirTemp("", "jpoet-*")
	if err != nil {
		return "", err
	}

	err = jpoet.Eval(
		jpoet.FileImport([]string{}),
		jpoet.FSImport(lib),
		jpoet.FSImport(imports.Fs),
		jpoet.StringImport("input/files.json", string(b)),
		jpoet.WithPlugin(markdown.Plugin()),
		jpoet.WithPlugin(jsonnet.Plugin()),
		jpoet.TLACode("files", "import 'input/files.json'"),
		jpoet.FileInput("./lib/manifest.libsonnet"),
		jpoet.Serialize(false),
		jpoet.DirectoryOutput(buildDir),
	)
	if err != nil {
		return "", err
	}
	return buildDir, nil
}
