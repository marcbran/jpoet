package jpoet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

type Option func(*evalConfig)

type evalConfig struct {
	vmOpts  []func(*jsonnet.VM)
	closers []io.Closer

	importer CompoundImporter
	contents map[string]jsonnet.Contents

	nodeInput    *ast.Node
	snippetInput *snippetInput
	fileInput    *string

	writerOutput    io.Writer
	valueOutput     any
	directoryOutput string

	serializedFormat bool

	errs []error
}

type snippetInput struct {
	filename string
	snippet  string
}

func newEvalConfig() *evalConfig {
	return &evalConfig{
		contents:         make(map[string]jsonnet.Contents),
		writerOutput:     os.Stdout,
		serializedFormat: true,
	}
}

func Eval(opts ...Option) error {
	c := newEvalConfig()
	for _, opt := range opts {
		opt(c)
	}
	return c.eval()
}

func TLAVar(key, val string) Option {
	return func(c *evalConfig) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLAVar(key, val) })
	}
}

func TLACode(key, val string) Option {
	return func(c *evalConfig) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLACode(key, val) })
	}
}

func TLANode(key string, node ast.Node) Option {
	return func(c *evalConfig) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLANode(key, node) })
	}
}

func NodeInput(node ast.Node) Option {
	return func(c *evalConfig) {
		c.nodeInput = &node
		c.snippetInput = nil
		c.fileInput = nil
	}
}

func SnippetInput(filename, snippet string) Option {
	return func(c *evalConfig) {
		c.nodeInput = nil
		c.snippetInput = &snippetInput{filename, snippet}
		c.fileInput = nil
	}
}

func FileInput(filename string) Option {
	return func(c *evalConfig) {
		c.nodeInput = nil
		c.snippetInput = nil
		c.fileInput = &filename
	}
}

func Importer(i jsonnet.Importer) Option {
	return func(c *evalConfig) {
		c.importer.Importers = append(c.importer.Importers, i)
	}
}

func FileImport(jpaths []string) Option {
	return Importer(&jsonnet.FileImporter{JPaths: jpaths})
}

func FSImport(f fs.FS) Option {
	return Importer(&FSImporter{Fs: f})
}

func StringImport(filename, value string) Option {
	return func(c *evalConfig) {
		c.contents[filename] = jsonnet.MakeContents(value)
	}
}

func WithNativeFunction(f *jsonnet.NativeFunction) Option {
	return func(c *evalConfig) {
		if f == nil {
			return
		}
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.NativeFunction(f) })
	}
}

func WithPlugin(p *Plugin) Option {
	return func(c *evalConfig) {
		c.closers = append(c.closers, p)
		WithNativeFunction(p.NativeFunction())(c)
	}
}

func WithPluginSet(plugins ...*Plugin) Option {
	return func(c *evalConfig) {
		for _, p := range plugins {
			WithPlugin(p)(c)
		}
	}
}

func WriterOutput(w io.Writer) Option {
	return func(c *evalConfig) {
		c.writerOutput = w
		c.valueOutput = nil
		c.directoryOutput = ""
	}
}

func ValueOutput(out any) Option {
	return func(c *evalConfig) {
		c.writerOutput = nil
		c.valueOutput = out
		c.directoryOutput = ""
	}
}

func DirectoryOutput(dir string) Option {
	return func(c *evalConfig) {
		c.writerOutput = nil
		c.valueOutput = nil
		c.directoryOutput = dir
	}
}

func Serialize(s bool) Option {
	return func(c *evalConfig) {
		c.serializedFormat = s
	}
}

func (c *evalConfig) hasInput() bool {
	return c.nodeInput != nil || c.snippetInput != nil || c.fileInput != nil
}

func (c *evalConfig) error() error {
	if len(c.errs) == 0 {
		return nil
	}
	if len(c.errs) == 1 {
		return fmt.Errorf("failed to evaluate Jsonnet: %w", c.errs[0])
	}
	return fmt.Errorf("failed to evaluate Jsonnet: %s", c.errs)
}

func (c *evalConfig) eval() error {
	defer func() {
		for _, closer := range c.closers {
			err := closer.Close()
			if err != nil {
				c.errs = append(c.errs, err)
			}
		}
	}()
	if !c.hasInput() {
		c.errs = append(c.errs, errors.New("missing input"))
		return c.error()
	}
	if len(c.contents) > 0 {
		c.importer.Importers = append(c.importer.Importers, &MemoryImporter{
			Data: c.contents,
		})
	}
	vm := jsonnet.MakeVM()
	for _, opt := range c.vmOpts {
		opt(vm)
	}
	if len(c.importer.Importers) > 0 {
		vm.Importer(c.importer)
	}

	var serializedJson string
	var err error
	if c.nodeInput != nil {
		serializedJson, err = vm.Evaluate(*c.nodeInput)
		if err != nil {
			c.errs = append(c.errs, err)
			return c.error()
		}
	} else if c.snippetInput != nil {
		serializedJson, err = vm.EvaluateAnonymousSnippet(c.snippetInput.filename, c.snippetInput.snippet)
		if err != nil {
			c.errs = append(c.errs, err)
			return c.error()
		}
	} else if c.fileInput != nil {
		serializedJson, err = vm.EvaluateFile(*c.fileInput)
		if err != nil {
			c.errs = append(c.errs, err)
			return c.error()
		}
	}

	if c.writerOutput != nil {
		output := serializedJson
		if !c.serializedFormat {
			err := json.Unmarshal([]byte(serializedJson), &output)
			if err != nil {
				c.errs = append(c.errs, err)
				return c.error()
			}
		}
		_, err := c.writerOutput.Write([]byte(output))
		if err != nil {
			c.errs = append(c.errs, err)
			return c.error()
		}
	} else if c.valueOutput != nil {
		if c.serializedFormat {
			c.valueOutput = serializedJson
		} else {
			err := json.Unmarshal([]byte(serializedJson), c.valueOutput)
			if err != nil {
				c.errs = append(c.errs, err)
				return c.error()
			}
		}
	} else if c.directoryOutput != "" {
		var entries map[string]any
		err = json.Unmarshal([]byte(serializedJson), &entries)
		if err != nil {
			c.errs = append(c.errs, err)
			return c.error()
		}
		err = writeEntries(c.directoryOutput, entries, c.serializedFormat)
		if err != nil {
			c.errs = append(c.errs, err)
			return c.error()
		}
	}
	return c.error()
}

func writeEntries(directory string, entries map[string]any, serialized bool) error {
	for filename, c := range entries {
		switch content := c.(type) {
		case map[string]any:
			err := writeEntries(filepath.Join(directory, filename), content, serialized)
			if err != nil {
				return err
			}
		default:
			err := writeFile(filepath.Join(directory, filename), content, serialized)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(filename string, content any, serialized bool) error {
	var fileContent []byte
	if serialized {
		var err error
		fileContent, err = json.Marshal(content)
		if err != nil {
			return err
		}
	} else {
		stringContent, ok := content.(string)
		if !ok {
			return fmt.Errorf("expect string when writing output to file: %s, but got %T", filename, content)
		}
		fileContent = []byte(stringContent)
	}

	_, err := os.Stat(filename)

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err == nil {
		existingContent, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if bytes.Equal(existingContent, fileContent) {
			return nil
		}
	}

	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, fileContent, 0666)
	if err != nil {
		return err
	}
	return nil
}
