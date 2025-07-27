package jpoet

import (
	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/internal/plugin"
)

func Serve(functions []jsonnet.NativeFunction) {
	plugin.NewServer(functions).Serve()
}
