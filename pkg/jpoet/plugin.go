package jpoet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/internal/plugin"
)

type Plugin struct {
	name             string
	invoker          plugin.Invoker
	closer           io.Closer
	middleware       []Middleware
	actionMiddleware []ActionMiddleware
	watchSource      WatchSource
	actions          map[string]ActionFunc
}

type WatchSource interface {
	InvocationKey(funcName string, args []any) InvocationKey
	Acquire(key InvocationKey) (cancel func(), err error)
	SetChanges(changes func(keys []InvocationKey))
}

type InvocationKey string

type ActionFunc func(context.Context, map[string]any) (string, error)

type PluginOption func(*Plugin)

func NewPlugin(name string, functions []jsonnet.NativeFunction, opts ...PluginOption) *Plugin {
	p := &Plugin{
		name:    name,
		invoker: plugin.NewLocalInvoker(functions),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func NewClientPlugin(name string, path string, opts ...PluginOption) (*Plugin, error) {
	invoker, err := plugin.NewClientInvoker(name, path)
	if err != nil {
		return nil, err
	}
	p := &Plugin{
		name:    name,
		invoker: invoker,
		closer:  invoker,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func NewPluginsDir(pluginsDir string, opts ...PluginOption) ([]*Plugin, error) {
	entries, err := readPluginEntries(pluginsDir)
	if err != nil {
		return nil, err
	}
	var plugins []*Plugin
	for _, entry := range entries {
		name := entry.Name()
		p, err := NewClientPlugin(name, filepath.Join(pluginsDir, name, name), opts...)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func readPluginEntries(pluginsDir string) ([]os.DirEntry, error) {
	_, err := os.Stat(pluginsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return os.ReadDir(pluginsDir)
}

func (p *Plugin) WithOptions(opts ...PluginOption) *Plugin {
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type Invoker = plugin.Invoker

type Middleware func(Invoker) Invoker

func WithMiddleware(middleware ...Middleware) PluginOption {
	return func(p *Plugin) {
		p.middleware = append(p.middleware, middleware...)
	}
}

type InvokeHook func(next Invoker, funcName string, args []any) (any, error)

type hookInvoker struct {
	next Invoker
	hook InvokeHook
}

func (h hookInvoker) Invoke(funcName string, args []any) (any, error) {
	return h.hook(h.next, funcName, args)
}

func HookMiddleware(hook InvokeHook) Middleware {
	return func(next Invoker) Invoker {
		return hookInvoker{next: next, hook: hook}
	}
}

func WithHook(hook InvokeHook) PluginOption {
	return WithMiddleware(HookMiddleware(hook))
}

type Executor interface {
	Exec(ctx context.Context, action string, data map[string]any) (string, error)
}

type ActionMiddleware func(Executor) Executor

func WithActionMiddleware(middleware ...ActionMiddleware) PluginOption {
	return func(p *Plugin) {
		p.actionMiddleware = append(p.actionMiddleware, middleware...)
	}
}

type ActionHook func(ctx context.Context, next Executor, action string, data map[string]any) (string, error)

type hookExecutor struct {
	next Executor
	hook ActionHook
}

func (h hookExecutor) Exec(ctx context.Context, action string, data map[string]any) (string, error) {
	return h.hook(ctx, h.next, action, data)
}

func ActionHookMiddleware(hook ActionHook) ActionMiddleware {
	return func(next Executor) Executor {
		return hookExecutor{next: next, hook: hook}
	}
}

func WithActionHook(hook ActionHook) PluginOption {
	return WithActionMiddleware(ActionHookMiddleware(hook))
}

type actionTable map[string]ActionFunc

func (t actionTable) Exec(ctx context.Context, action string, data map[string]any) (string, error) {
	fn, ok := t[action]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrNoSuchAction, action)
	}
	return fn(ctx, data)
}

type multiCloser []io.Closer

func (m multiCloser) Close() error {
	var errs []error
	for _, c := range m {
		if c == nil {
			continue
		}
		err := c.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func WithCloser(closer io.Closer) PluginOption {
	return func(p *Plugin) {
		if p.closer != nil {
			p.closer = multiCloser{p.closer, closer}
			return
		}
		p.closer = closer
	}
}

func WithWatchSource(source WatchSource) PluginOption {
	return func(p *Plugin) {
		p.watchSource = source
	}
}

func (p *Plugin) WatchSource() WatchSource {
	return p.watchSource
}

func WithAction(name string, fn ActionFunc) PluginOption {
	return func(p *Plugin) {
		if p.actions == nil {
			p.actions = make(map[string]ActionFunc)
		}
		p.actions[name] = fn
	}
}

func (p *Plugin) ActionNames() []string {
	names := make([]string, 0, len(p.actions))
	for name := range p.actions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (p *Plugin) Action(name string) (ActionFunc, bool) {
	if _, ok := p.actions[name]; !ok {
		return nil, false
	}
	executor := Executor(actionTable(p.actions))
	for _, m := range p.actionMiddleware {
		executor = m(executor)
	}
	return func(ctx context.Context, data map[string]any) (string, error) {
		return executor.Exec(ctx, name, data)
	}, true
}

func (p *Plugin) Name() string {
	return p.name
}

func (p *Plugin) Serve() {
	plugin.NewConsumer(p.name, p.invoker).Serve()
}

func (p *Plugin) NativeFunction() *jsonnet.NativeFunction {
	invoker := plugin.Invoker(p.invoker)
	for _, m := range p.middleware {
		invoker = m(invoker)
	}
	return plugin.NewConsumer(p.name, invoker).Function()
}

func (p *Plugin) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}
