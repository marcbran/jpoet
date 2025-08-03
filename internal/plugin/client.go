package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/marcbran/jpoet/internal/plugin/proto"
	"os/exec"
	"path/filepath"
	"strings"
)

type Client struct {
	NamedInvoker
	client *plugin.Client
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
		NamedInvoker: NamedInvoker{
			Invoker: invoker,
			name:    name,
		},
		client: client,
	}, nil
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
