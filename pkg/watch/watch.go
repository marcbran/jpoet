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

	dirtyMu          sync.Mutex
	dirtyInvocations map[pluginInvocation]struct{}
	dirtyPaths       map[string]struct{}
	dirtyNotify      chan struct{}

	dirWatcher *directoryWatcher

	lifecycle *jpoet.Lifecycle
}

func New(env *jpoet.Environment) (*Environment, error) {
	dirWatcher, err := newDirectoryWatcher()
	if err != nil {
		return nil, err
	}

	invocations := newInvocationRegistry()
	we := &Environment{
		Environment: env,

		watches:     newWatchRegistry(invocations),
		invocations: invocations,

		dirtyInvocations: map[pluginInvocation]struct{}{},
		dirtyPaths:       map[string]struct{}{},
		dirtyNotify:      make(chan struct{}, 1),

		dirWatcher: dirWatcher,
	}
	we.lifecycle = jpoet.NewLifecycle(func() error {
		return errors.Join(env.Close(), dirWatcher.Close())
	})

	for _, p := range env.Plugins() {
		source := p.WatchSource()
		if source == nil {
			continue
		}
		source.SetChanges(we.changesFunc(p))
	}

	we.lifecycle.Go(we.runDirWatch)
	we.lifecycle.Go(we.runDirtyDispatch)
	we.lifecycle.Go(we.runEviction)
	return we, nil
}

func (we *Environment) Close() error {
	err := we.lifecycle.Close()
	we.watches.close()
	we.invocations.close()
	return err
}

func (we *Environment) Eval(opts ...jpoet.EvalOption) error {
	var invocations []jpoet.Invocation
	var paths []string
	err := we.Environment.Eval(append(opts, jpoet.EvalInvocations(&invocations), jpoet.EvalImportedPaths(&paths))...)
	we.touchInvocations(toInvocations(invocations))
	we.dirWatcher.track(paths)
	return err
}

type Result struct {
	Output string
	Err    error
}

func (we *Environment) Watch(opts ...WatchOption) (func(), error) {
	c := watchConfig{}
	for _, opt := range opts {
		opt(&c)
	}
	if c.key == "" {
		return nil, errors.New("watch key is required")
	}
	if !c.hasInput() {
		return nil, errors.New("missing input")
	}
	if !c.hasOutput() {
		return nil, errors.New("watch output is required")
	}

	var s sink
	var ch chan Result
	if c.valueOutput != nil {
		ch = make(chan Result, 1)
		s = &channelSink{ch: ch}
	} else {
		s = fileSink{path: c.fileOutput}
	}

	we.watches.ensure(c.key, c, s)
	initial, _, _ := we.evaluate(c.key)

	if c.valueOutput != nil {
		*c.valueOutput.initial = initial
		*c.valueOutput.updates = ch
	} else {
		s.deliver(initial)
	}

	unregister := we.unregisterFunc(c.key, s)
	if initial.Err != nil {
		return unregister, initial.Err
	}
	return unregister, nil
}

func (we *Environment) evalForWatch(c watchConfig) (Result, []pluginInvocation) {
	serialize := true
	if c.serialize != nil {
		serialize = *c.serialize
	}
	opts := []jpoet.EvalOption{jpoet.EvalSerialize(serialize)}
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
	var paths []string
	opts = append(opts, jpoet.EvalImportedPaths(&paths))

	err := we.Environment.Eval(opts...)
	invocations := toInvocations(recorded)
	we.touchInvocations(invocations)
	we.dirWatcher.track(paths)
	if err != nil {
		return Result{Err: err}, invocations
	}
	return Result{Output: out.String()}, invocations
}

func (we *Environment) unregisterFunc(key WatchKey, s sink) func() {
	return func() {
		we.watches.detach(key, s)
		s.close()
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
			we.markDirtyInvocation(pluginInvocation{key: key, plugin: p})
		}
	}
}

func (we *Environment) runDirWatch() {
	we.dirWatcher.run(we.lifecycle.Done(), we.markDirtyPath)
}

func (we *Environment) markDirtyInvocation(inv pluginInvocation) {
	we.dirtyMu.Lock()
	_, exists := we.dirtyInvocations[inv]
	we.dirtyInvocations[inv] = struct{}{}
	we.dirtyMu.Unlock()
	if exists {
		return
	}
	we.notifyDirty()
}

func (we *Environment) markDirtyPath(path string) {
	we.dirtyMu.Lock()
	_, exists := we.dirtyPaths[path]
	we.dirtyPaths[path] = struct{}{}
	we.dirtyMu.Unlock()
	if exists {
		return
	}
	we.notifyDirty()
}

func (we *Environment) notifyDirty() {
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
		invocations, paths := we.drainDirty()
		if we.handleDirtyPaths(paths) {
			continue
		}
		we.handleDirtyInvocations(invocations)
	}
}

func (we *Environment) drainDirty() (map[pluginInvocation]struct{}, map[string]struct{}) {
	we.dirtyMu.Lock()
	defer we.dirtyMu.Unlock()
	invocations := we.dirtyInvocations
	paths := we.dirtyPaths
	we.dirtyInvocations = map[pluginInvocation]struct{}{}
	we.dirtyPaths = map[string]struct{}{}
	return invocations, paths
}

func (we *Environment) handleDirtyInvocations(invocations map[pluginInvocation]struct{}) {
	dirtyKeys := map[WatchKey]struct{}{}
	for inv := range invocations {
		for _, key := range we.watches.matchingInvocation(inv) {
			dirtyKeys[key] = struct{}{}
		}
	}
	for key := range dirtyKeys {
		if we.watches.hasSubscribers(key) {
			we.reevaluate(key)
		}
	}
}

func (we *Environment) handleDirtyPaths(paths map[string]struct{}) bool {
	if len(paths) == 0 {
		return false
	}
	we.FlushCache()
	for _, key := range we.watches.subscribedKeys() {
		we.reevaluate(key)
	}
	return true
}

func (we *Environment) reevaluate(key WatchKey) Result {
	result, subs, changed := we.evaluate(key)
	if !changed {
		return result
	}
	for _, s := range subs {
		s.deliver(result)
	}
	return result
}

func (we *Environment) evaluate(key WatchKey) (Result, []sink, bool) {
	input, ok := we.watches.input(key)
	if !ok {
		return Result{}, nil, false
	}

	result, invocations := we.evalForWatch(input)
	subs, changed := we.watches.update(key, result, invocations)
	return result, subs, changed
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
	serialize    *bool

	valueOutput *valueOutput
	fileOutput  string
}

type valueOutput struct {
	initial *Result
	updates *<-chan Result
}

func (c *watchConfig) hasInput() bool {
	return c.nodeInput != nil || c.snippetInput != nil || c.fileInput != nil
}

func (c *watchConfig) hasOutput() bool {
	return c.valueOutput != nil || c.fileOutput != ""
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

func WatchSerialize(s bool) WatchOption {
	return func(c *watchConfig) {
		c.serialize = &s
	}
}

func WatchValueOutput(initial *Result, updates *<-chan Result) WatchOption {
	return func(c *watchConfig) {
		c.valueOutput = &valueOutput{initial: initial, updates: updates}
		c.fileOutput = ""
	}
}

func WatchFileOutput(path string) WatchOption {
	return func(c *watchConfig) {
		c.fileOutput = path
		c.valueOutput = nil
	}
}
