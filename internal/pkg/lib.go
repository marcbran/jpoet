package pkg

import (
	"embed"
	"errors"
	"github.com/marcbran/jpoet/internal/pkg/lib/imports"
	"github.com/marcbran/jpoet/pkg/jpoet"
	"os"
	"path/filepath"
)

//go:embed lib
var lib embed.FS

type Config struct {
	Source      string      `json:"source"`
	Description string      `json:"description"`
	Coordinates Coordinates `json:"coordinates"`
	Usage       Usage       `json:"usage"`
	Plugins     []Plugin    `json:"plugins"`
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
	mainFile := filepath.Join(pkgDir, "main.libsonnet")
	mainCode := []byte("{}")
	_, err := os.Stat(mainFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
	} else {
		mainCode, err = os.ReadFile(mainFile)
		if err != nil {
			return Config{}, err
		}
	}

	pkgFile := filepath.Join(pkgDir, "pkg.libsonnet")
	_, err = os.Stat(pkgFile)
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
	err = jpoet.NewEval().
		FSImport(lib).
		FSImport(imports.Fs).
		StringImport("input/main.libsonnet", string(mainCode)).
		StringImport("input/pkg.libsonnet", string(pkgCode)).
		TLACode("lib", "import 'input/main.libsonnet'").
		TLACode("pkg", "import 'input/pkg.libsonnet'").
		TLACode("examples", "null").
		TLACode("examplesString", "null").
		FileInput("./lib/resolve_pkg_config.libsonnet").
		Serialize(false).
		ValueOutput(&config).
		Eval()
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
