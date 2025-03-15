package fun

import (
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/formatter"
)

func FormatJsonnet() *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   "formatJsonnet",
		Params: ast.Identifiers{"code"},
		Func: func(input []any) (any, error) {
			if len(input) != 1 {
				return nil, fmt.Errorf("code must be provided")
			}
			code, ok := input[0].(string)
			if !ok {
				return nil, fmt.Errorf("code must be a string")
			}
			formattedCode, err := formatter.Format("main.jsonnet", code, formatter.DefaultOptions())
			if err != nil {
				return nil, err
			}
			return formattedCode, nil
		},
	}
}
