package jpoet

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func Eval(opts ...Option) error {
	c := newConfig()
	for _, opt := range opts {
		opt(c)
	}
	return c.eval()
}

type Option func(*config)

type config struct {
	envConfig
	evalConfig

	closers []io.Closer
	errs    []error
}

func newConfig() *config {
	return &config{
		envConfig:  newEnvConfig(),
		evalConfig: newEvalConfig(),
	}
}

func (c *config) eval() error {
	defer func() {
		for _, closer := range c.closers {
			err := closer.Close()
			if err != nil {
				c.errs = append(c.errs, err)
			}
		}
	}()
	vm := c.buildVM()
	err := c.run(vm, c.recorder)
	if err != nil {
		c.errs = append(c.errs, err)
	}
	return c.error()
}

func (c *config) error() error {
	if len(c.errs) == 0 {
		return nil
	}
	if len(c.errs) == 1 {
		return fmt.Errorf("failed to evaluate Jsonnet: %w", c.errs[0])
	}
	return fmt.Errorf("failed to evaluate Jsonnet: %s", c.errs)
}

func TLAVar(key, val string) Option {
	return func(c *config) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLAVar(key, val) })
	}
}

func TLACode(key, val string) Option {
	return func(c *config) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLACode(key, val) })
	}
}

func TLANode(key string, node ast.Node) Option {
	return func(c *config) {
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.TLANode(key, node) })
	}
}

func NodeInput(node ast.Node) Option {
	return func(c *config) {
		c.nodeInput = &node
		c.snippetInput = nil
		c.fileInput = nil
	}
}

func SnippetInput(filename, snippet string) Option {
	return func(c *config) {
		c.nodeInput = nil
		c.snippetInput = &snippetInput{filename, snippet}
		c.fileInput = nil
	}
}

func FileInput(filename string) Option {
	return func(c *config) {
		c.nodeInput = nil
		c.snippetInput = nil
		c.fileInput = &filename
	}
}

func Importer(i jsonnet.Importer) Option {
	return func(c *config) {
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
	return func(c *config) {
		c.contents[filename] = jsonnet.MakeContents(value)
	}
}

func WithNativeFunction(f *jsonnet.NativeFunction) Option {
	return func(c *config) {
		if f == nil {
			return
		}
		c.vmOpts = append(c.vmOpts, func(vm *jsonnet.VM) { vm.NativeFunction(f) })
	}
}

func WithPlugin(p *Plugin) Option {
	return func(c *config) {
		c.closers = append(c.closers, p)
		nf := p.recordingNativeFunction(c.recorder)
		WithNativeFunction(nf)(c)
	}
}

func WithPluginSet(plugins ...*Plugin) Option {
	return func(c *config) {
		for _, p := range plugins {
			WithPlugin(p)(c)
		}
	}
}

func WriterOutput(w io.Writer) Option {
	return func(c *config) {
		c.writerOutput = w
		c.valueOutput = nil
		c.directoryOutput = ""
	}
}

func ValueOutput(out any) Option {
	return func(c *config) {
		c.writerOutput = nil
		c.valueOutput = out
		c.directoryOutput = ""
	}
}

func DirectoryOutput(dir string) Option {
	return func(c *config) {
		c.writerOutput = nil
		c.valueOutput = nil
		c.directoryOutput = dir
	}
}

func Serialize(s bool) Option {
	return func(c *config) {
		c.serializedFormat = s
	}
}
