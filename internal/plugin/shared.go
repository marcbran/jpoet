package plugin

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/marcbran/jpoet/internal/plugin/proto"
	"google.golang.org/grpc"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "JSONNET_PLUGIN",
	MagicCookieValue: "9af0e0b1-a9c4-47ed-a8c2-fc740428f447",
}

type Invoker interface {
	Invoke(funcName string, args []any) (any, error)
}

type InvokeArgs struct {
	FuncName string
	Args     []any
}

type grpcPlugin struct {
	plugin.Plugin
	Impl Invoker
}

func (p *grpcPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterInvokerServer(s, &grpcServerInvoker{impl: p.Impl})
	return nil
}

func (p *grpcPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
	return &grpcClientInvoker{client: proto.NewInvokerClient(c)}, nil
}
