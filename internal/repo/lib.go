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

	err = jpoet.NewEval().
		FileImport([]string{}).
		FSImport(lib).
		FSImport(imports.Fs).
		StringImport("input/files.json", string(b)).
		Plugin(markdown.Plugin()).
		Plugin(jsonnet.Plugin()).
		TLACode("files", "import 'input/files.json'").
		FileInput("./lib/manifest.libsonnet").
		Serialize(false).
		DirectoryOutput(buildDir).
		Eval()
	if err != nil {
		return "", err
	}
	return buildDir, nil
}
