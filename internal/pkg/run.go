package pkg

import (
	"errors"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/internal/plugin"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func Run(filename string) error {
	vm := jsonnet.MakeVM()

	dir := path.Dir(filename)
	pluginsDir := filepath.Join(dir, ".jpoet", "plugins")

	var err error
	var errs []error
	var clients []*plugin.Client
	defer func() {
		if err != nil {
			errs = append(errs, err)
		}
		for _, client := range clients {
			err := client.Close()
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			err = errors.New(fmt.Sprint(errs))
		}
	}()

	entries, err := readEntries(pluginsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		client, err := plugin.NewClient(filepath.Join(pluginsDir, name, name))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if client == nil {
			continue
		}
		clients = append(clients, client)
		vm.NativeFunction(client.InvokeFunction())
	}

	output, err := vm.EvaluateFile(filename)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}

func readEntries(pluginsDir string) ([]os.DirEntry, error) {
	_, err := os.Stat(pluginsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}
	return entries, nil
}
