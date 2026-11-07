package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/marcbran/jpoet/pkg/jpoet"
)

func main() {
	jpoet.NewPlugin("fixture", nil, jpoet.WithFunctions(jpoet.Function{
		Name: "gate",
		Func: gate,
	})).Serve()
}

func gate(ctx context.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, errors.New("gate expects a directory")
	}
	dir, ok := args[0].(string)
	if !ok {
		return nil, errors.New("gate expects a directory")
	}
	err := os.WriteFile(filepath.Join(dir, "entered"), nil, 0o600)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		err = os.WriteFile(filepath.Join(dir, "cancelled"), nil, 0o600)
		if err != nil {
			return nil, err
		}
		return nil, ctx.Err()
	case <-time.After(time.Minute):
		return nil, errors.New("gate was never cancelled")
	}
}
