package plugin

import (
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
)

type Client struct {
	name   string
	path   string
	logger hclog.Logger
}

func NewClient(
	dir, name string,
) *Client {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	return &Client{
		name:   name,
		path:   filepath.Join(dir, ".jpoet", name, "plugin"),
		logger: logger,
	}
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
			return c.Invoke(funcName, args)
		},
	}
}

func (c Client) Invoke(funcName string, args []any) (any, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"invoker": &Plugin{},
		},
		Cmd:    exec.Command(c.path),
		Logger: c.logger,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense("invoker")
	if err != nil {
		return nil, err
	}
	inv := raw.(Invoker)

	return inv.Invoke(funcName, args)
}

type rpcClientInvoker struct {
	client *rpc.Client
}

func (c *rpcClientInvoker) Invoke(funcName string, args []any) (any, error) {
	var resp any
	err := c.client.Call("Plugin.Invoke", InvokeArgs{
		FuncName: funcName,
		Args:     args,
	}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}
