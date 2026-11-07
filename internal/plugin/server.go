package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-plugin"
	"github.com/marcbran/jpoet/internal/plugin/proto"
)

func (c Consumer) Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"invoker": &grpcPlugin{
				Impl: c.invoker,
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
	resp, err := s.impl.Invoke(ctx, request.FuncName, args)
	if f, ok := asFailure(err); ok {
		b, err := json.Marshal(f)
		if err != nil {
			return nil, err
		}
		return &proto.InvokeResponse{Failure: b}, nil
	}
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

type localInvoker struct {
	functionNames string
	functions     map[string]Function
}

func NewLocalInvoker(functions []Function) Invoker {
	functionMap := make(map[string]Function)
	var functionNames []string
	for _, f := range functions {
		functionNames = append(functionNames, f.Name)
		functionMap[f.Name] = f
	}
	sort.Strings(functionNames)
	return &localInvoker{
		functionNames: strings.Join(functionNames, ", "),
		functions:     functionMap,
	}
}

func NativeFunctions(natives []jsonnet.NativeFunction) []Function {
	functions := make([]Function, len(natives))
	for i, nf := range natives {
		inner := nf.Func
		functions[i] = Function{
			Name: nf.Name,
			Func: func(_ context.Context, args []any) (any, error) {
				return inner(args)
			},
		}
	}
	return functions
}

func (i localInvoker) Invoke(ctx context.Context, funcName string, args []any) (any, error) {
	f, ok := i.functions[funcName]
	if !ok {
		return "", fmt.Errorf("no such function: %s, available functions: %s", funcName, i.functionNames)
	}
	res, err := f.Func(ctx, args)
	if err != nil {
		return "", err
	}
	return res, nil
}
