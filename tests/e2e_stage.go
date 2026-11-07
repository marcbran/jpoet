//go:build e2e

package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jpoet/pkg/watch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const waitTimeout = 10 * time.Second
const quietPeriod = 200 * time.Millisecond

const prelude = `local t(f, n) = std.native('invoke:test')(f, [n]);
local fixture(f, a) = std.native('invoke:fixture')(f, [a]);
`

type Stage struct {
	t   *testing.T
	dir string

	test    *testPlugin
	plugins []*jpoet.Plugin

	env      *jpoet.Environment
	watchEnv *watch.Environment

	mu      sync.Mutex
	evals   map[string]*backgroundEval
	watches map[string]*watchSubscription

	flushDone chan struct{}
	closeDone chan struct{}
	poolSize  int
	gatedFS   *gatedFS
}

type backgroundEval struct {
	cancel context.CancelFunc
	done   chan struct{}
	output string
	err    error
}

type watchSubscription struct {
	cancel     context.CancelFunc
	ready      chan struct{}
	initial    watch.Result
	updates    <-chan watch.Result
	unregister func()
	err        error
}

func scenario(t *testing.T) (*Stage, *Stage, *Stage) {
	s := &Stage{
		t:        t,
		dir:      t.TempDir(),
		test:     newTestPlugin(),
		evals:    map[string]*backgroundEval{},
		watches:  map[string]*watchSubscription{},
		poolSize: 4,
	}
	s.plugins = []*jpoet.Plugin{s.test.plugin()}
	t.Cleanup(s.cleanup)
	return s, s, s
}

func (s *Stage) cleanup() {
	s.test.mu.Lock()
	gates := make([]*gate, 0, len(s.test.gates))
	for _, g := range s.test.gates {
		gates = append(gates, g)
	}
	s.test.mu.Unlock()
	for _, g := range gates {
		g.open()
	}
	if s.gatedFS != nil {
		s.gatedFS.open()
	}
	s.mu.Lock()
	for _, e := range s.evals {
		e.cancel()
	}
	for _, w := range s.watches {
		w.cancel()
	}
	s.mu.Unlock()
	for _, p := range s.plugins {
		_ = p.Close()
	}
	if s.watchEnv != nil {
		_ = s.watchEnv.Close()
	}
	if s.env != nil {
		_ = s.env.Close()
	}
}

func (s *Stage) and() *Stage {
	return s
}

func (s *Stage) the_fixture_plugin() *Stage {
	pluginPath := filepath.Join(s.dir, "plugins", "fixture", "jsonnet-plugin-fixture")
	cmd := exec.Command("go", "build", "-o", pluginPath, "./fixture")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(s.t, cmd.Run(), stderr.String())
	p, err := jpoet.NewClientPlugin("fixture", pluginPath)
	require.NoError(s.t, err)
	s.plugins = append(s.plugins, p)
	return s
}

func (s *Stage) a_library_file(name string, content string) *Stage {
	require.NoError(s.t, os.WriteFile(filepath.Join(s.dir, name), []byte(content), 0o600))
	return s
}

func (s *Stage) the_library_file_changes(name string, content string) *Stage {
	return s.a_library_file(name, content)
}

func (s *Stage) a_closed_gate(name string) *Stage {
	s.test.gate(name).close()
	return s
}

func (s *Stage) an_open_gate(name string, value string) *Stage {
	g := s.test.gate(name)
	g.set(value)
	g.open()
	return s
}

func (s *Stage) an_environment_with_pool_size(size int) *Stage {
	s.poolSize = size
	return s.an_environment()
}

func (s *Stage) a_watched_environment_with_pool_size(size int) *Stage {
	s.poolSize = size
	return s.a_watched_environment()
}

func (s *Stage) an_environment() *Stage {
	s.env = jpoet.Env(s.envOptions()...)
	return s
}

func (s *Stage) a_watched_environment() *Stage {
	env := jpoet.Env(s.envOptions()...)
	we, err := watch.New(env)
	require.NoError(s.t, err)
	s.watchEnv = we
	s.env = env
	return s
}

func (s *Stage) envOptions() []jpoet.EnvOption {
	opts := []jpoet.EnvOption{
		jpoet.EnvFileImport([]string{s.dir}),
		jpoet.EnvWithPluginSet(s.plugins...),
		jpoet.EnvPoolSize(s.poolSize),
	}
	if s.gatedFS != nil {
		opts = append(opts, jpoet.EnvFSImport(s.gatedFS))
	}
	return opts
}

func (s *Stage) a_gated_library_file(name string, content string) *Stage {
	s.gatedFS = newGatedFS(name, content)
	return s
}

func (s *Stage) the_library_read_is_entered() *Stage {
	select {
	case <-s.gatedFS.entered:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "library read was not entered")
	}
	return s
}

func (s *Stage) the_library_read_is_released() *Stage {
	s.gatedFS.open()
	return s
}

func (s *Stage) an_evaluation_is_started(name string, snippet string) *Stage {
	ctx, cancel := context.WithCancel(context.Background())
	return s.startEvaluation(ctx, cancel, name, snippet)
}

func (s *Stage) an_already_cancelled_evaluation_is_started(name string, snippet string) *Stage {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return s.startEvaluation(ctx, cancel, name, snippet)
}

func (s *Stage) startEvaluation(ctx context.Context, cancel context.CancelFunc, name string, snippet string) *Stage {
	e := &backgroundEval{cancel: cancel, done: make(chan struct{})}
	s.mu.Lock()
	s.evals[name] = e
	s.mu.Unlock()
	go func() {
		defer close(e.done)
		var out bytes.Buffer
		e.err = s.evaluator()(ctx,
			jpoet.EvalSnippetInput(filepath.Join(s.dir, name+".jsonnet"), prelude+snippet),
			jpoet.EvalWriterOutput(&out),
		)
		e.output = out.String()
	}()
	return s
}

func (s *Stage) a_fixture_gate_evaluation_is_started(name string) *Stage {
	return s.an_evaluation_is_started(name, fmt.Sprintf("fixture('gate', %q)", s.dir))
}

func (s *Stage) evaluator() func(context.Context, ...jpoet.EvalOption) error {
	if s.watchEnv != nil {
		return s.watchEnv.Eval
	}
	return s.env.Eval
}

func (s *Stage) an_evaluation_runs(name string, snippet string) *Stage {
	s.an_evaluation_is_started(name, snippet)
	return s.the_evaluation_finishes(name)
}

func (s *Stage) the_evaluation_is_cancelled(name string) *Stage {
	s.eval(name).cancel()
	return s
}

func (s *Stage) the_evaluation_finishes(name string) *Stage {
	select {
	case <-s.eval(name).done:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "evaluation did not finish", name)
	}
	return s
}

func (s *Stage) the_evaluation_is_still_running(name string) *Stage {
	select {
	case <-s.eval(name).done:
		require.FailNow(s.t, "evaluation finished early", name)
	default:
	}
	return s
}

func (s *Stage) the_evaluation_outputs(name string, expected string) *Stage {
	e := s.eval(name)
	require.NoError(s.t, e.err)
	assert.JSONEq(s.t, expected, e.output)
	return s
}

func (s *Stage) the_evaluation_fails_with_cancellation(name string) *Stage {
	assert.ErrorIs(s.t, s.eval(name).err, context.Canceled)
	return s
}

func (s *Stage) eval(name string) *backgroundEval {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.evals[name]
	require.True(s.t, ok, "unknown evaluation %s", name)
	return e
}

func (s *Stage) the_gate_is_entered(name string) *Stage {
	select {
	case <-s.test.gate(name).entered:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "gate was not entered", name)
	}
	return s
}

func (s *Stage) the_gate_is_opened(name string) *Stage {
	s.test.gate(name).open()
	return s
}

func (s *Stage) the_gate_is_closed(name string) *Stage {
	s.test.gate(name).close()
	return s
}

func (s *Stage) the_gate_value_becomes(name string, value string) *Stage {
	s.test.gate(name).set(value)
	return s
}

func (s *Stage) the_gate_observes_cancellation(name string) *Stage {
	select {
	case <-s.test.gate(name).cancelled:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "gate did not observe cancellation", name)
	}
	return s
}

func (s *Stage) the_count_is(name string, expected int) *Stage {
	assert.Equal(s.t, expected, s.test.count(name))
	return s
}

func (s *Stage) the_fixture_gate_is_entered() *Stage {
	s.fileAppears("entered")
	return s
}

func (s *Stage) the_fixture_gate_observes_cancellation() *Stage {
	s.fileAppears("cancelled")
	return s
}

func (s *Stage) fileAppears(name string) {
	path := filepath.Join(s.dir, name)
	require.Eventually(s.t, func() bool {
		_, err := os.Stat(path)
		return err == nil
	}, waitTimeout, 10*time.Millisecond, "%s did not appear", path)
}

func (s *Stage) the_cache_is_flushed() *Stage {
	s.env.FlushCache()
	return s
}

func (s *Stage) the_watched_environment_is_closed_in_the_background() *Stage {
	s.closeDone = make(chan struct{})
	go func() {
		defer close(s.closeDone)
		_ = s.watchEnv.Close()
	}()
	return s
}

func (s *Stage) the_close_finishes() *Stage {
	select {
	case <-s.closeDone:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "close did not finish")
	}
	return s
}

func (s *Stage) the_cache_is_flushed_in_the_background() *Stage {
	s.flushDone = make(chan struct{})
	go func() {
		defer close(s.flushDone)
		s.env.FlushCache()
	}()
	return s
}

func (s *Stage) the_flush_is_pending() *Stage {
	select {
	case <-s.flushDone:
		require.FailNow(s.t, "flush finished while an evaluation was running")
	case <-time.After(quietPeriod):
	}
	return s
}

func (s *Stage) the_flush_finishes() *Stage {
	select {
	case <-s.flushDone:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "flush did not finish")
	}
	return s
}

func (s *Stage) a_watch_is_started(key string, snippet string) *Stage {
	ctx, cancel := context.WithCancel(context.Background())
	return s.startWatch(ctx, cancel, key, key, snippet)
}

func (s *Stage) a_subscriber_joins(name string, key string, snippet string) *Stage {
	ctx, cancel := context.WithCancel(context.Background())
	return s.startWatch(ctx, cancel, name, key, snippet)
}

func (s *Stage) an_already_cancelled_subscriber_joins(name string, key string, snippet string) *Stage {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return s.startWatch(ctx, cancel, name, key, snippet)
}

func (s *Stage) startWatch(ctx context.Context, cancel context.CancelFunc, name string, key string, snippet string) *Stage {
	w := &watchSubscription{cancel: cancel, ready: make(chan struct{})}
	s.mu.Lock()
	s.watches[name] = w
	s.mu.Unlock()
	go func() {
		defer close(w.ready)
		w.unregister, w.err = s.watchEnv.Watch(ctx,
			watch.WatchSnippetInput(filepath.Join(s.dir, key+".jsonnet"), prelude+snippet),
			watch.WatchWithKey(watch.WatchKey(key)),
			watch.WatchValueOutput(&w.initial, &w.updates),
		)
	}()
	return s
}

func (s *Stage) the_watch_is_established(key string) *Stage {
	w := s.watch(key)
	select {
	case <-w.ready:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "watch was not established", key)
	}
	require.NoError(s.t, w.err)
	return s
}

func (s *Stage) the_subscription_fails_with_cancellation(name string) *Stage {
	w := s.watch(name)
	select {
	case <-w.ready:
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "subscription did not finish", name)
	}
	assert.ErrorIs(s.t, w.err, context.Canceled)
	return s
}

func (s *Stage) the_initial_watch_output_is(key string, expected string) *Stage {
	w := s.watch(key)
	require.NoError(s.t, w.initial.Err)
	assert.JSONEq(s.t, expected, w.initial.Output)
	return s
}

func (s *Stage) the_invocation_changes(name string) *Stage {
	s.test.change(name)
	return s
}

func (s *Stage) the_watch_receives(key string, expected string) *Stage {
	select {
	case r, ok := <-s.watch(key).updates:
		require.True(s.t, ok, "watch %s was closed", key)
		require.NoError(s.t, r.Err)
		assert.JSONEq(s.t, expected, r.Output)
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "watch received nothing", key)
	}
	return s
}

func (s *Stage) the_watch_receives_nothing(key string) *Stage {
	select {
	case r, ok := <-s.watch(key).updates:
		if ok {
			require.FailNow(s.t, "watch received an unexpected update", "%s: %+v", key, r)
		}
	case <-time.After(quietPeriod):
	}
	return s
}

func (s *Stage) the_watch_is_unsubscribed(key string) *Stage {
	s.watch(key).unregister()
	return s
}

func (s *Stage) the_watch_is_closed_without_updates(key string) *Stage {
	select {
	case r, ok := <-s.watch(key).updates:
		require.False(s.t, ok, "watch %s received %+v", key, r)
	case <-time.After(waitTimeout):
		require.FailNow(s.t, "watch was not closed", key)
	}
	return s
}

func (s *Stage) watch(key string) *watchSubscription {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.watches[key]
	require.True(s.t, ok, "unknown watch %s", key)
	return w
}
