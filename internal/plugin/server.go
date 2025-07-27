package plugin

import (
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"sort"
	"strings"
)

type Server struct {
	functions []jsonnet.NativeFunction
	logger    hclog.Logger
}

func NewServer(functions []jsonnet.NativeFunction) *Server {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})
	return &Server{
		functions: functions,
		logger:    logger,
	}
}

func (s Server) Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"invoker": &Plugin{
				Impl: newFunctionInvoker(s.functions, s.logger),
			},
		},
		Logger: s.logger,
	})
}

type rpcServerInvoker struct {
	Impl Invoker
}

func (s *rpcServerInvoker) Invoke(args InvokeArgs, resp *any) error {
	var err error
	*resp, err = s.Impl.Invoke(args.FuncName, args.Args)
	if err != nil {
		return err
	}
	return nil
}

type functionInvoker struct {
	functionNames string
	functions     map[string]jsonnet.NativeFunction
	logger        hclog.Logger
}

func newFunctionInvoker(
	functions []jsonnet.NativeFunction,
	logger hclog.Logger,
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
		logger:        logger,
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
