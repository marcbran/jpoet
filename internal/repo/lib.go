package repo

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/marcbran/devsonnet/internal/pkg/lib/imports"
	"github.com/marcbran/gensonnet/pkg/gensonnet"
	"github.com/marcbran/gensonnet/pkg/gensonnet/config"
	"os"
)

//go:embed lib
var lib embed.FS

func manifestRepo(ctx context.Context, files map[string]string) (string, error) {
	b, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	buildDir, err := os.MkdirTemp("", "devsonnet-*")
	if err != nil {
		return "", err
	}

	err = gensonnet.RenderWithConfig(ctx, config.Config{
		Render: config.RenderConfig{
			TargetDir: buildDir,
			Lib: config.LibConfig{
				ManifestCode: `
					local files = import 'input/files.json';

					local manifest = import 'lib/manifest.libsonnet';
					manifest(files)
				`,
				Filesystems: []embed.FS{
					lib,
					imports.Fs,
				},
				Imports: map[string]string{
					"input/files.json": string(b),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}
	return buildDir, nil
}
