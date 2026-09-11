package jpoet

import (
	"github.com/google/go-jsonnet"
)

type Invocation struct {
	Key    InvocationKey
	Plugin *Plugin
}

type invocationRecorder struct {
	invocations []Invocation
}

func (r *invocationRecorder) record(inv Invocation) {
	if r == nil {
		return
	}
	r.invocations = append(r.invocations, inv)
}

func (r *invocationRecorder) take() []Invocation {
	if r == nil {
		return nil
	}
	invocations := r.invocations
	r.invocations = nil
	return invocations
}

func (p *Plugin) recordingNativeFunction(invocations *invocationRecorder) *jsonnet.NativeFunction {
	nf := p.NativeFunction()
	inner := nf.Func
	wrapped := *nf
	wrapped.Func = func(input []any) (any, error) {
		result, err := inner(input)
		if err == nil && len(input) == 2 && p.watchSource != nil {
			funcName, _ := input[0].(string)
			args, _ := input[1].([]any)
			invocations.record(Invocation{Key: p.watchSource.InvocationKey(funcName, args), Plugin: p})
		}
		return result, err
	}
	return &wrapped
}

type pathRecorder struct {
	paths []string
}

func (r *pathRecorder) record(path string) {
	if r == nil {
		return
	}
	r.paths = append(r.paths, path)
}

func (r *pathRecorder) take() []string {
	if r == nil {
		return nil
	}
	paths := r.paths
	r.paths = nil
	return paths
}

type recordingImporter struct {
	inner    jsonnet.Importer
	recorder *pathRecorder
}

func (r *recordingImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	contents, foundAt, err := r.inner.Import(importedFrom, importedPath)
	if err == nil {
		r.recorder.record(foundAt)
	}
	return contents, foundAt, err
}
