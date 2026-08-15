package pkg

import (
	"embed"
	"errors"
	"os"
	"path/filepath"

	"github.com/marcbran/jpoet/internal/pkg/lib/imports"
	"github.com/marcbran/jpoet/pkg/jpoet"
)

//go:embed lib
var lib embed.FS

type Config struct {
	Source      string      `json:"source"`
	Description string      `json:"description"`
	Coordinates Coordinates `json:"coordinates"`
	Usage       Usage       `json:"usage"`
	Plugins     []Plugin    `json:"plugins"`
	External    []string    `json:"external"`
}

type Coordinates struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
	Repo   string `json:"repo"`
}

type Usage struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type Plugin struct {
	Github *GithubPlugin `json:"github"`
}

type GithubPlugin struct {
	Repo    string `json:"repo"`
	Version string `json:"version"`
}

func ResolvePkgConfig(pkgDir string) (Config, error) {
	pkgFile := filepath.Join(pkgDir, "pkg.libsonnet")
	_, err := os.Stat(pkgFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		return Config{}, errors.New("pkg.libsonnet not found")
	}
	pkgCode, err := os.ReadFile(pkgFile)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = jpoet.Eval(
		jpoet.FSImport(lib),
		jpoet.FSImport(imports.Fs),
		jpoet.StringImport("input/pkg.libsonnet", string(pkgCode)),
		jpoet.SnippetInput("pkg.libsonnet", "import 'input/pkg.libsonnet'"),
		jpoet.Serialize(false),
		jpoet.ValueOutput(&config),
	)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
