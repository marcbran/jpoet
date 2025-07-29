package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-plugin"
	"github.com/marcbran/jpoet/internal/plugin/proto"
	"sort"
	"strings"
)

type Server struct {
	functions []jsonnet.NativeFunction
}

func NewServer(
	functions []jsonnet.NativeFunction,
) *Server {
	return &Server{
		functions: functions,
	}
}

func (s Server) Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"invoker": &grpcPlugin{
				Impl: newFunctionInvoker(s.functions),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     newLogger(),
	})
}

type grpcServerInvoker struct {
	proto.UnimplementedInvokerServer
	impl Invoker
}

func (s grpcServerInvoker) Invoke(
	ctx context.Context,
	request *proto.InvokeRequest,
) (*proto.InvokeResponse, error) {
	var args []any
	err := json.Unmarshal(request.Args, &args)
	if err != nil {
		return nil, err
	}
	resp, err := s.impl.Invoke(request.FuncName, args)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return &proto.InvokeResponse{
		Value: b,
	}, nil
}

type functionInvoker struct {
	functionNames string
	functions     map[string]jsonnet.NativeFunction
}

func newFunctionInvoker(
	functions []jsonnet.NativeFunction,
) *functionInvoker {
	functionMap := make(map[string]jsonnet.NativeFunction)
	var functionNames []string
	for _, f := range functions {
		functionNames = append(functionNames, f.Name)
		functionMap[f.Name] = f
	}
	sort.Strings(functionNames)
	return &functionInvoker{
		functionNames: strings.Join(functionNames, ", "),
		functions:     functionMap,
	}
}

func (i *functionInvoker) Invoke(funcName string, args []any) (any, error) {
	f, ok := i.functions[funcName]
	if !ok {
		return "", fmt.Errorf("no such function: %s, available functions: %s", funcName, i.functionNames)
	}
	res, err := f.Func(args)
	if err != nil {
		return "", err
	}
	return res, nil
}
