package jpoet

import (
	"context"

	"github.com/google/go-jsonnet"
)

type Invocation struct {
	Key    InvocationKey
	Plugin *Plugin
}

type evalState struct {
	ctx         context.Context
	invocations []Invocation
	paths       []string
}

type evalSlot struct {
	state *evalState
}

func (s *evalSlot) context() context.Context {
	return s.state.ctx
}

func (s *evalSlot) recordInvocation(inv Invocation) {
	s.state.invocations = append(s.state.invocations, inv)
}

func (s *evalSlot) recordPath(path string) {
	s.state.paths = append(s.state.paths, path)
}

func (p *Plugin) recordingNativeFunction(slot *evalSlot) *jsonnet.NativeFunction {
	nf := p.nativeFunction(slot.context)
	inner := nf.Func
	wrapped := *nf
	wrapped.Func = func(input []any) (any, error) {
		result, err := inner(input)
		if err == nil && len(input) == 2 && p.watchSource != nil {
			funcName, _ := input[0].(string)
			args, _ := input[1].([]any)
			slot.recordInvocation(Invocation{Key: p.watchSource.InvocationKey(funcName, args), Plugin: p})
		}
		return result, err
	}
	return &wrapped
}

type recordingImporter struct {
	inner jsonnet.Importer
	slot  *evalSlot
}

func (r *recordingImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	contents, foundAt, err := r.inner.Import(importedFrom, importedPath)
	if err == nil {
		r.slot.recordPath(foundAt)
	}
	return contents, foundAt, err
}
