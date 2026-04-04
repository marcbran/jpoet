package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/marcbran/jpoet/internal/plugin/proto"
)

type client struct {
	invoker Invoker
	client  *plugin.Client
}

func NewClientInvoker(
	name string,
	path string,
) (InvokeCloser, error) {
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "jsonnet-plugin-") {
		return nil, fmt.Errorf("plugin path does not start with jsonnet-plugin")
	}
	baseName := strings.TrimPrefix(base, "jsonnet-plugin-")
	if baseName != name {
		return nil, fmt.Errorf("plugin name does not match path")
	}

	pluginClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"invoker": &grpcPlugin{},
		},
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           newLogger(),
	})

	rpcClient, err := pluginClient.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense("invoker")
	if err != nil {
		return nil, err
	}
	invoker := raw.(Invoker)

	return &client{
		invoker: invoker,
		client:  pluginClient,
	}, nil
}

func (c *client) Invoke(funcName string, args []any) (any, error) {
	return c.invoker.Invoke(funcName, args)
}

func (c *client) Close() error {
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
