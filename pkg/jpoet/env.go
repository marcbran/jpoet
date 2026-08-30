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

type Environment struct {
	vm      *jsonnet.VM
	closers []io.Closer
}

func (e *Environment) Eval(opts ...EvalOption) error {
	c := newEvalConfig()
	for _, opt := range opts {
		opt(&c)
	}
	return c.run(e.vm)
}

func (e *Environment) Close() error {
	var errs []error
	for _, closer := range e.closers {
		err := closer.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type EnvOption func(*envConfig)

type envConfig struct {
	vmOpts  []func(*jsonnet.VM)
	closers []io.Closer

	importer CompoundImporter
	contents map[string]jsonnet.Contents
}

func newEnvConfig() envConfig {
	return envConfig{contents: make(map[string]jsonnet.Contents)}
}

func (c *envConfig) buildVM() *jsonnet.VM {
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
	return vm
}

func Env(opts ...EnvOption) *Environment {
	c := newEnvConfig()
	for _, opt := range opts {
		opt(&c)
	}
	return &Environment{vm: c.buildVM(), closers: c.closers}
}

func EnvTLAVar(key, val string) EnvOption {
	return func(c *envConfig) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLAVar(key, val) })
	}
}

func EnvTLACode(key, val string) EnvOption {
	return func(c *envConfig) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLACode(key, val) })
	}
}

func EnvTLANode(key string, node ast.Node) EnvOption {
	return func(c *envConfig) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLANode(key, node) })
	}
}

func EnvImporter(i jsonnet.Importer) EnvOption {
	return func(c *envConfig) {
		c.importer.Importers = append(c.importer.Importers, i)
	}
}

func EnvFileImport(jpaths []string) EnvOption {
	return EnvImporter(&jsonnet.FileImporter{JPaths: jpaths})
}

func EnvFSImport(f fs.FS) EnvOption {
	return EnvImporter(&FSImporter{Fs: f})
}

func EnvStringImport(filename, value string) EnvOption {
	return func(c *envConfig) {
		c.contents[filename] = jsonnet.MakeContents(value)
	}
}

func EnvWithNativeFunction(f *jsonnet.NativeFunction) EnvOption {
	return func(c *envConfig) {
		if f == nil {
			return
		}
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.NativeFunction(f) })
	}
}

func EnvWithPlugin(p *Plugin) EnvOption {
	return func(c *envConfig) {
		c.closers = append(c.closers, p)
		EnvWithNativeFunction(p.NativeFunction())(c)
	}
}

func EnvWithPluginSet(plugins ...*Plugin) EnvOption {
	return func(c *envConfig) {
		for _, p := range plugins {
			EnvWithPlugin(p)(c)
		}
	}
}

type EvalOption func(*evalConfig)

type evalConfig struct {
	nodeInput    *ast.Node
	snippetInput *snippetInput
	fileInput    *string

	writerOutput    io.Writer
	valueOutput     any
	directoryOutput string

	serializedFormat bool
}

type snippetInput struct {
	filename string
	snippet  string
}

func newEvalConfig() evalConfig {
	return evalConfig{
		writerOutput:     os.Stdout,
		serializedFormat: true,
	}
}

func (c *evalConfig) hasInput() bool {
	return c.nodeInput != nil || c.snippetInput != nil || c.fileInput != nil
}

func (c *evalConfig) run(vm *jsonnet.VM) error {
	if !c.hasInput() {
		return errors.New("missing input")
	}
	serializedJson, err := evaluateInput(vm, c.nodeInput, c.snippetInput, c.fileInput)
	if err != nil {
		return err
	}
	return writeOutput(serializedJson, c.writerOutput, c.valueOutput, c.directoryOutput, c.serializedFormat)
}

func evaluateInput(vm *jsonnet.VM, nodeInput *ast.Node, snippetInput *snippetInput, fileInput *string) (string, error) {
	if nodeInput != nil {
		return vm.Evaluate(*nodeInput)
	}
	if snippetInput != nil {
		return vm.EvaluateAnonymousSnippet(snippetInput.filename, snippetInput.snippet)
	}
	return vm.EvaluateFile(*fileInput)
}

func writeOutput(serializedJson string, writerOutput io.Writer, valueOutput any, directoryOutput string, serializedFormat bool) error {
	if writerOutput != nil {
		output := serializedJson
		if !serializedFormat {
			err := json.Unmarshal([]byte(serializedJson), &output)
			if err != nil {
				return err
			}
		}
		_, err := writerOutput.Write([]byte(output))
		return err
	}
	if valueOutput != nil {
		if serializedFormat {
			return nil
		}
		return json.Unmarshal([]byte(serializedJson), valueOutput)
	}
	if directoryOutput != "" {
		var entries map[string]any
		err := json.Unmarshal([]byte(serializedJson), &entries)
		if err != nil {
			return err
		}
		return writeEntries(directoryOutput, entries, serializedFormat)
	}
	return nil
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

func EvalNodeInput(node ast.Node) EvalOption {
	return func(c *evalConfig) {
		c.nodeInput = &node
		c.snippetInput = nil
		c.fileInput = nil
	}
}

func EvalSnippetInput(filename, snippet string) EvalOption {
	return func(c *evalConfig) {
		c.nodeInput = nil
		c.snippetInput = &snippetInput{filename, snippet}
		c.fileInput = nil
	}
}

func EvalFileInput(filename string) EvalOption {
	return func(c *evalConfig) {
		c.nodeInput = nil
		c.snippetInput = nil
		c.fileInput = &filename
	}
}

func EvalWriterOutput(w io.Writer) EvalOption {
	return func(c *evalConfig) {
		c.writerOutput = w
		c.valueOutput = nil
		c.directoryOutput = ""
	}
}

func EvalValueOutput(out any) EvalOption {
	return func(c *evalConfig) {
		c.writerOutput = nil
		c.valueOutput = out
		c.directoryOutput = ""
	}
}

func EvalDirectoryOutput(dir string) EvalOption {
	return func(c *evalConfig) {
		c.writerOutput = nil
		c.valueOutput = nil
		c.directoryOutput = dir
	}
}

func EvalSerialize(s bool) EvalOption {
	return func(c *evalConfig) {
		c.serializedFormat = s
	}
}
