package watch

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/google/go-jsonnet/ast"
	"github.com/marcbran/jpoet/pkg/jpoet"
)

const watchMaxIdle = 5 * time.Minute
const invocationMaxIdle = 10 * time.Minute

type InvocationKey = jpoet.InvocationKey

type WatchSource = jpoet.WatchSource

type pluginInvocation struct {
	key    InvocationKey
	plugin *jpoet.Plugin
}

type Environment struct {
	*jpoet.Environment

	watches     *watchRegistry
	invocations *invocationRegistry

	dirtyMu     sync.Mutex
	dirty       map[pluginInvocation]struct{}
	dirtyNotify chan struct{}

	lifecycle *jpoet.Lifecycle
}

func New(env *jpoet.Environment) *Environment {
	invocations := newInvocationRegistry()
	we := &Environment{
		Environment: env,
		watches:     newWatchRegistry(invocations),
		invocations: invocations,
		dirty:       map[pluginInvocation]struct{}{},
		dirtyNotify: make(chan struct{}, 1),
		lifecycle:   jpoet.NewLifecycle(env.Close),
	}

	hasWatchSupport := false
	for _, p := range env.Plugins() {
		source := p.WatchSource()
		if source == nil {
			continue
		}
		hasWatchSupport = true
		source.SetChanges(we.changesFunc(p))
	}

	we.lifecycle.Go(we.runEviction)
	if hasWatchSupport {
		we.lifecycle.Go(we.runDirtyDispatch)
	}
	return we
}

func (we *Environment) Close() error {
	err := we.lifecycle.Close()
	we.invocations.close()
	return err
}

func (we *Environment) Eval(opts ...jpoet.EvalOption) error {
	var invocations []jpoet.Invocation
	err := we.Environment.Eval(append(opts, jpoet.EvalInvocations(&invocations))...)
	we.touchInvocations(toInvocations(invocations))
	return err
}

func (we *Environment) Watch(opts ...WatchOption) (string, <-chan string, func(), error) {
	c := watchConfig{}
	for _, opt := range opts {
		opt(&c)
	}
	if c.key == "" {
		return "", nil, nil, errors.New("watch key is required")
	}
	if !c.hasInput() {
		return "", nil, nil, errors.New("missing input")
	}

	if initial, ch, ok := we.watches.attach(c.key); ok {
		return initial, ch, we.unregisterFunc(c.key, ch), nil
	}

	output, invocations, err := we.evalForWatch(c)
	if err != nil {
		return "", nil, nil, err
	}

	initial, ch := we.watches.create(c.key, c, output, invocations)

	return initial, ch, we.unregisterFunc(c.key, ch), nil
}

func (we *Environment) evalForWatch(c watchConfig) (string, []pluginInvocation, error) {
	opts := []jpoet.EvalOption{jpoet.EvalSerialize(true)}
	switch {
	case c.nodeInput != nil:
		opts = append(opts, jpoet.EvalNodeInput(*c.nodeInput))
	case c.snippetInput != nil:
		opts = append(opts, jpoet.EvalSnippetInput(c.snippetInput.filename, c.snippetInput.snippet))
	case c.fileInput != nil:
		opts = append(opts, jpoet.EvalFileInput(*c.fileInput))
	}
	var out bytes.Buffer
	opts = append(opts, jpoet.EvalWriterOutput(&out))
	var recorded []jpoet.Invocation
	opts = append(opts, jpoet.EvalInvocations(&recorded))

	err := we.Environment.Eval(opts...)
	invocations := toInvocations(recorded)
	we.touchInvocations(invocations)
	if err != nil {
		return "", nil, err
	}
	return out.String(), invocations, nil
}

func (we *Environment) unregisterFunc(key WatchKey, ch chan string) func() {
	return func() {
		we.watches.detach(key, ch)
	}
}

func (we *Environment) touchInvocations(invocations []pluginInvocation) {
	for _, inv := range invocations {
		we.invocations.touch(inv)
	}
}

func toInvocations(invocations []jpoet.Invocation) []pluginInvocation {
	out := make([]pluginInvocation, len(invocations))
	for i, inv := range invocations {
		out[i] = pluginInvocation{key: inv.Key, plugin: inv.Plugin}
	}
	return out
}

func (we *Environment) changesFunc(p *jpoet.Plugin) func(keys []InvocationKey) {
	return func(keys []InvocationKey) {
		for _, key := range keys {
			we.markDirty(pluginInvocation{key: key, plugin: p})
		}
	}
}

func (we *Environment) markDirty(inv pluginInvocation) {
	we.dirtyMu.Lock()
	_, exists := we.dirty[inv]
	we.dirty[inv] = struct{}{}
	we.dirtyMu.Unlock()
	if exists {
		return
	}
	select {
	case we.dirtyNotify <- struct{}{}:
	default:
	}
}

func (we *Environment) runDirtyDispatch() {
	for {
		select {
		case <-we.lifecycle.Done():
			return
		case <-we.dirtyNotify:
		}
		for {
			inv, ok := we.popDirty()
			if !ok {
				break
			}
			we.handleChange(inv)
		}
	}
}

func (we *Environment) popDirty() (pluginInvocation, bool) {
	we.dirtyMu.Lock()
	defer we.dirtyMu.Unlock()
	for inv := range we.dirty {
		delete(we.dirty, inv)
		return inv, true
	}
	return pluginInvocation{}, false
}

func (we *Environment) handleChange(change pluginInvocation) {
	for _, key := range we.watches.matching(change) {
		we.reevaluate(key)
	}
}

func (we *Environment) reevaluate(key WatchKey) {
	input, ok := we.watches.input(key)
	if !ok {
		return
	}

	output, invocations, err := we.evalForWatch(input)
	if err != nil {
		return
	}

	subs := we.watches.update(key, output, invocations)
	for _, ch := range subs {
		select {
		case ch <- output:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- output:
			default:
			}
		}
	}
}

func (we *Environment) runEviction() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-we.lifecycle.Done():
			return
		case <-ticker.C:
			we.watches.evictIdle(watchMaxIdle)
			we.invocations.evictIdle(invocationMaxIdle)
		}
	}
}

type WatchKey string

type WatchOption func(*watchConfig)

type snippetInput struct {
	filename string
	snippet  string
}

type watchConfig struct {
	nodeInput    *ast.Node
	snippetInput *snippetInput
	fileInput    *string
	key          WatchKey
}

func (c *watchConfig) hasInput() bool {
	return c.nodeInput != nil || c.snippetInput != nil || c.fileInput != nil
}

func WatchNodeInput(node ast.Node) WatchOption {
	return func(c *watchConfig) {
		c.nodeInput = &node
		c.snippetInput = nil
		c.fileInput = nil
	}
}

func WatchSnippetInput(filename, snippet string) WatchOption {
	return func(c *watchConfig) {
		c.nodeInput = nil
		c.snippetInput = &snippetInput{filename, snippet}
		c.fileInput = nil
	}
}

func WatchFileInput(filename string) WatchOption {
	return func(c *watchConfig) {
		c.nodeInput = nil
		c.snippetInput = nil
		c.fileInput = &filename
	}
}

func WatchWithKey(key WatchKey) WatchOption {
	return func(c *watchConfig) {
		c.key = key
	}
}
