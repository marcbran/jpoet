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

func (p *Plugin) recordingNativeFunction(recorder *invocationRecorder) *jsonnet.NativeFunction {
	nf := p.NativeFunction()
	inner := nf.Func
	wrapped := *nf
	wrapped.Func = func(input []any) (any, error) {
		result, err := inner(input)
		if err == nil && len(input) == 2 && p.watchSource != nil {
			funcName, _ := input[0].(string)
			args, _ := input[1].([]any)
			recorder.record(Invocation{Key: p.watchSource.InvocationKey(funcName, args), Plugin: p})
		}
		return result, err
	}
	return &wrapped
}
