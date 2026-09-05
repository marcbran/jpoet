package jpoet

import (
	"errors"
	"io/fs"
	"sync"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var ErrEnvironmentClosed = errors.New("environment is closed")

func Env(opts ...EnvOption) *Environment {
	c := newEnvConfig()
	for _, opt := range opts {
		opt(&c)
	}
	e := &Environment{
		plugins:  c.plugins,
		recorder: c.recorder,
		vm:       c.buildVM(),
	}
	e.lifecycle = NewLifecycle(nil)
	return e
}

type Environment struct {
	plugins  []*Plugin
	recorder *invocationRecorder

	vm   *jsonnet.VM
	vmMu sync.Mutex

	lifecycle *Lifecycle
}

func (e *Environment) Plugins() []*Plugin {
	return e.plugins
}

func (e *Environment) Close() error {
	return e.lifecycle.Close()
}

type EnvOption func(*envConfig)

type envConfig struct {
	vmOpts  []func(*jsonnet.VM)
	plugins []*Plugin

	importer CompoundImporter
	contents map[string]jsonnet.Contents
	recorder *invocationRecorder
}

func newEnvConfig() envConfig {
	return envConfig{contents: make(map[string]jsonnet.Contents), recorder: &invocationRecorder{}}
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
		c.plugins = append(c.plugins, p)
		nf := p.recordingNativeFunction(c.recorder)
		EnvWithNativeFunction(nf)(c)
	}
}

func EnvWithPluginSet(plugins ...*Plugin) EnvOption {
	return func(c *envConfig) {
		for _, p := range plugins {
			EnvWithPlugin(p)(c)
		}
	}
}
