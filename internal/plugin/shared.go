package plugin

import (
	"context"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
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

type NamedInvoker struct {
	Invoker
	name string
}

func (i NamedInvoker) Function() *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   fmt.Sprintf("invoke:%s", i.name),
		Params: ast.Identifiers{"funcName", "args"},
		Func: func(input []any) (any, error) {
			if len(input) != 2 {
				return nil, fmt.Errorf("funcName and args must be provided")
			}
			funcName, ok := input[0].(string)
			if !ok {
				return nil, fmt.Errorf("funcName must be a string")
			}
			args, ok := input[1].([]any)
			if !ok {
				return nil, fmt.Errorf("args must be an array")
			}
			return i.Invoke(funcName, args)
		},
	}
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
