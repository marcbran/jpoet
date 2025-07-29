package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/hashicorp/go-plugin"
	"github.com/marcbran/jpoet/internal/plugin/proto"
	"os/exec"
	"path/filepath"
	"strings"
)

type Client struct {
	name    string
	client  *plugin.Client
	invoker Invoker
}

func NewClient(path string) (*Client, error) {
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "jsonnet-plugin-") {
		return nil, fmt.Errorf("plugin path does not start with jsonnet-plugin")
	}
	name := strings.TrimPrefix(base, "jsonnet-plugin-")

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"invoker": &grpcPlugin{},
		},
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           newLogger(),
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense("invoker")
	if err != nil {
		return nil, err
	}
	invoker := raw.(Invoker)

	return &Client{
		name:    name,
		client:  client,
		invoker: invoker,
	}, nil
}

func (c Client) InvokeFunction() *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   fmt.Sprintf("invoke:%s", c.name),
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
			return c.invoker.Invoke(funcName, args)
		},
	}
}

func (c Client) Close() error {
	c.client.Kill()
	return nil
}

type grpcClientInvoker struct {
	client proto.InvokerClient
}

func (c grpcClientInvoker) Invoke(funcName string, args []any) (any, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Invoke(context.Background(), &proto.InvokeRequest{
		FuncName: funcName,
		Args:     b,
	})
	if err != nil {
		return nil, err
	}
	var res any
	err = json.Unmarshal(resp.Value, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
