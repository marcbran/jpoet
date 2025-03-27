package manifest

import (
	"embed"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/marcbran/jsonnet-kit/internal/manifest/fun"
	"github.com/marcbran/jsonnet-kit/pkg/jsonnext"
	"os"
	"path"
)

//go:embed lib
var lib embed.FS

func RunDir(dirname string, jpath []string) error {
	vm := jsonnet.MakeVM()
	vm.MaxStack = 1000000
	vm.Importer(jsonnext.CompoundImporter{
		Importers: []jsonnet.Importer{
			&jsonnext.FSImporter{Fs: lib},
			&jsonnet.FileImporter{
				JPaths: jpath,
			},
		},
	})
	vm.NativeFunction(fun.FormatJsonnet())
	vm.TLACode("manifest", fmt.Sprintf("import '%s/main.jsonnet'", dirname))
	vm.StringOutput = true

	files, err := vm.EvaluateFileMulti("lib/main.libsonnet")
	if err != nil {
		return err
	}
	for name, content := range files {
		filename := path.Join(dirname, name)
		err := os.MkdirAll(path.Dir(filename), 0755)
		if err != nil {
			return err
		}
		err = os.WriteFile(filename, []byte(content), 0666)
		if err != nil {
			return err
		}
	}
	return nil
}
