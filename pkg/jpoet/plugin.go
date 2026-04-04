package jpoet

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/internal/plugin"
)

type Invoker = plugin.Invoker

type Middleware func(Invoker) Invoker

type Plugin struct {
	name       string
	invoker    plugin.Invoker
	closer     io.Closer
	middleware []Middleware
}

func NewPlugin(name string, functions []jsonnet.NativeFunction) *Plugin {
	return &Plugin{
		name:    name,
		invoker: plugin.NewLocalInvoker(functions),
	}
}

func NewClientPlugin(name string, path string) (*Plugin, error) {
	invoker, err := plugin.NewClientInvoker(name, path)
	if err != nil {
		return nil, err
	}
	return &Plugin{
		name:    name,
		invoker: invoker,
		closer:  invoker,
	}, nil
}

func NewPluginsDir(pluginsDir string, middleware ...Middleware) ([]*Plugin, error) {
	entries, err := readPluginEntries(pluginsDir)
	if err != nil {
		return nil, err
	}
	var plugins []*Plugin
	for _, entry := range entries {
		name := entry.Name()
		p, err := NewClientPlugin(name, filepath.Join(pluginsDir, name, name))
		if err != nil {
			return nil, err
		}
		if len(middleware) > 0 {
			p = p.WithMiddleware(middleware...)
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func (p *Plugin) WithMiddleware(middleware ...Middleware) *Plugin {
	return &Plugin{
		name:       p.name,
		invoker:    p.invoker,
		closer:     p.closer,
		middleware: append(p.middleware, middleware...),
	}
}

func (p *Plugin) Serve() {
	plugin.NewConsumer(p.name, p.invoker).Serve()
}

func (p *Plugin) NativeFunction() *jsonnet.NativeFunction {
	invoker := plugin.Invoker(p.invoker)
	for _, m := range p.middleware {
		invoker = m(invoker)
	}
	return plugin.NewConsumer(p.name, invoker).Function()
}

func (p *Plugin) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}

func readPluginEntries(pluginsDir string) ([]os.DirEntry, error) {
	_, err := os.Stat(pluginsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return os.ReadDir(pluginsDir)
}
