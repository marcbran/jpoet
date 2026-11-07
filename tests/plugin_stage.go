//go:build e2e

package tests

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"sync"
	"testing/fstest"

	"github.com/marcbran/jpoet/pkg/jpoet"
)

type gate struct {
	mu      sync.Mutex
	closed  bool
	value   string
	release chan struct{}

	entered   chan struct{}
	cancelled chan struct{}
}

func newGate(value string) *gate {
	return &gate{
		value:     value,
		release:   make(chan struct{}),
		entered:   make(chan struct{}, 16),
		cancelled: make(chan struct{}, 16),
	}
}

func (g *gate) close() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return
	}
	g.closed = true
	g.release = make(chan struct{})
}

func (g *gate) open() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.closed {
		return
	}
	g.closed = false
	close(g.release)
}

func (g *gate) set(value string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

func (g *gate) current() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

type cancellation int

const (
	failOnCancel cancellation = iota
	ignoreCancel
	failureOnCancel
)

func (g *gate) pass(ctx context.Context, onCancel cancellation) (any, error) {
	g.mu.Lock()
	closed, release := g.closed, g.release
	g.mu.Unlock()
	if !closed {
		return g.current(), nil
	}
	g.entered <- struct{}{}
	if onCancel == ignoreCancel {
		<-release
		return g.current(), nil
	}
	select {
	case <-release:
		return g.current(), nil
	case <-ctx.Done():
		g.cancelled <- struct{}{}
		if onCancel == failureOnCancel {
			return nil, jpoet.NewFailure(http.StatusServiceUnavailable, ctx.Err().Error())
		}
		return nil, ctx.Err()
	}
}

type gatedFS struct {
	files   fstest.MapFS
	release chan struct{}
	entered chan struct{}
}

func newGatedFS(name string, content string) *gatedFS {
	return &gatedFS{
		files:   fstest.MapFS{name: &fstest.MapFile{Data: []byte(content)}},
		release: make(chan struct{}),
		entered: make(chan struct{}, 16),
	}
}

func (f *gatedFS) Open(name string) (fs.File, error) {
	return f.files.Open(name)
}

func (f *gatedFS) ReadFile(name string) ([]byte, error) {
	data, err := f.files.ReadFile(name)
	if err != nil {
		return nil, err
	}
	f.entered <- struct{}{}
	<-f.release
	return data, nil
}

func (f *gatedFS) open() {
	select {
	case <-f.release:
	default:
		close(f.release)
	}
}

type testPlugin struct {
	mu      sync.Mutex
	gates   map[string]*gate
	counts  map[string]int
	changes func(keys []jpoet.InvocationKey)
}

func newTestPlugin() *testPlugin {
	return &testPlugin{gates: map[string]*gate{}, counts: map[string]int{}}
}

func (p *testPlugin) gate(name string) *gate {
	p.mu.Lock()
	defer p.mu.Unlock()
	g, ok := p.gates[name]
	if !ok {
		g = newGate(name)
		p.gates[name] = g
	}
	return g
}

func (p *testPlugin) count(name string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.counts[name]
}

func (p *testPlugin) change(name string) {
	p.mu.Lock()
	changes := p.changes
	p.mu.Unlock()
	changes([]jpoet.InvocationKey{jpoet.InvocationKey(name)})
}

func (p *testPlugin) plugin() *jpoet.Plugin {
	return jpoet.NewPlugin("test", nil,
		jpoet.WithFunctions(
			jpoet.Function{Name: "gate", Func: func(ctx context.Context, args []any) (any, error) {
				name, err := nameArg(args)
				if err != nil {
					return nil, err
				}
				return p.gate(name).pass(ctx, failOnCancel)
			}},
			jpoet.Function{Name: "soft", Func: func(ctx context.Context, args []any) (any, error) {
				name, err := nameArg(args)
				if err != nil {
					return nil, err
				}
				return p.gate(name).pass(ctx, failureOnCancel)
			}},
			jpoet.Function{Name: "hold", Func: func(ctx context.Context, args []any) (any, error) {
				name, err := nameArg(args)
				if err != nil {
					return nil, err
				}
				return p.gate(name).pass(ctx, ignoreCancel)
			}},
			jpoet.Function{Name: "count", Func: func(ctx context.Context, args []any) (any, error) {
				name, err := nameArg(args)
				if err != nil {
					return nil, err
				}
				p.mu.Lock()
				defer p.mu.Unlock()
				p.counts[name]++
				return float64(p.counts[name]), nil
			}},
		),
		jpoet.WithWatchSource(p),
	)
}

func (p *testPlugin) InvocationKey(funcName string, args []any) jpoet.InvocationKey {
	name, _ := nameArg(args)
	return jpoet.InvocationKey(name)
}

func (p *testPlugin) Acquire(key jpoet.InvocationKey) (func(), error) {
	return func() {}, nil
}

func (p *testPlugin) SetChanges(changes func(keys []jpoet.InvocationKey)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.changes = changes
}

func nameArg(args []any) (string, error) {
	if len(args) != 1 {
		return "", errors.New("expected a single name argument")
	}
	name, ok := args[0].(string)
	if !ok {
		return "", errors.New("expected a string name argument")
	}
	return name, nil
}
