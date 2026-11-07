package jpoet

import (
	"context"
	"errors"
	"io/fs"
	"math"
	"runtime"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"golang.org/x/sync/semaphore"
)

var ErrEnvironmentClosed = errors.New("environment is closed")

func Env(opts ...EnvOption) *Environment {
	c := newEnvConfig()
	for _, opt := range opts {
		opt(&c)
	}
	importer := c.importer()
	e := &Environment{
		plugins:   c.plugins,
		importers: c.importers,
		pool: newVMPool(c.poolSize, func() *pooledVM {
			return c.buildVM(importer)
		}),
		flushBarrier: semaphore.NewWeighted(flushWeight),
	}
	e.lifecycle = NewLifecycle(nil)
	return e
}

type Environment struct {
	plugins   []*Plugin
	importers []jsonnet.Importer

	pool         *vmPool
	flushBarrier *semaphore.Weighted
	generation   uint64

	lifecycle *Lifecycle
}

func (e *Environment) Plugins() []*Plugin {
	return e.plugins
}

type flushableImporter interface {
	FlushCache()
}

const flushWeight = math.MaxInt32

func (e *Environment) FlushCache() {
	_ = e.flushBarrier.Acquire(context.Background(), flushWeight)
	defer e.flushBarrier.Release(flushWeight)
	for _, imp := range e.importers {
		if f, ok := imp.(flushableImporter); ok {
			f.FlushCache()
		}
	}
	e.generation++
}

func (e *Environment) Close() error {
	return e.lifecycle.Close()
}

type EnvOption func(*envConfig)

type envConfig struct {
	vmOpts    []func(*jsonnet.VM)
	plugins   []*Plugin
	importers []jsonnet.Importer
	contents  map[string]jsonnet.Contents
	poolSize  int
}

func newEnvConfig() envConfig {
	return envConfig{
		contents: make(map[string]jsonnet.Contents),
		poolSize: runtime.GOMAXPROCS(0),
	}
}

func (c *envConfig) importer() jsonnet.Importer {
	importers := c.importers
	if len(c.contents) > 0 {
		importers = append(importers, &MemoryImporter{Data: c.contents})
	}
	if len(importers) == 0 {
		return nil
	}
	return CompoundImporter{Importers: importers}
}

func (c *envConfig) buildVM(importer jsonnet.Importer) *pooledVM {
	vm := jsonnet.MakeVM()
	for _, opt := range c.vmOpts {
		opt(vm)
	}
	slot := &evalSlot{}
	for _, p := range c.plugins {
		vm.NativeFunction(p.recordingNativeFunction(slot))
	}
	v := &pooledVM{vm: vm, slot: slot}
	if importer != nil {
		v.importer = &recordingImporter{inner: importer, slot: slot}
		vm.Importer(v.importer)
	}
	return v
}

func EnvPoolSize(size int) EnvOption {
	return func(c *envConfig) {
		c.poolSize = size
	}
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
		c.importers = append(c.importers, i)
	}
}

func EnvFileImport(jpaths []string) EnvOption {
	return EnvImporter(&FileImporter{JPaths: jpaths})
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
		c.plugins = append(c.plugins, p)
	}
}

func EnvWithPluginSet(plugins ...*Plugin) EnvOption {
	return func(c *envConfig) {
		for _, p := range plugins {
			EnvWithPlugin(p)(c)
		}
	}
}
