package jpoet

import (
	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/internal/plugin"
)

type Plugin struct {
	server *plugin.Server
}

func NewPlugin(name string, functions []jsonnet.NativeFunction) *Plugin {
	return &Plugin{
		server: plugin.NewServer(name, functions),
	}
}

func (p *Plugin) Serve() {
	p.server.Serve()
}

func (p *Plugin) NativeFunction() *jsonnet.NativeFunction {
	return p.server.Function()
}
