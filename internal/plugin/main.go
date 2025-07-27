package plugin

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

type Plugin struct {
	Impl Invoker
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcServerInvoker{Impl: p.Impl}, nil
}

func (Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClientInvoker{client: c}, nil
}
