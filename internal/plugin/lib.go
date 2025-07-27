package plugin

import (
	"github.com/hashicorp/go-plugin"
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
