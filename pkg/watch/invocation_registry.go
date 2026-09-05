package watch

import (
	"sync"
	"time"
)

type invocationEntry struct {
	refs       int
	cancel     func()
	lastAccess time.Time
}

type invocationRegistry struct {
	mu      sync.Mutex
	entries map[InvocationKey]*invocationEntry
	pending map[InvocationKey]chan struct{}
}

func newInvocationRegistry() *invocationRegistry {
	return &invocationRegistry{
		entries: map[InvocationKey]*invocationEntry{},
		pending: map[InvocationKey]chan struct{}{},
	}
}

func (r *invocationRegistry) ref(inv pluginInvocation, delta int) {
	var entry *invocationEntry
	if delta > 0 {
		entry = r.ensure(inv)
	} else {
		r.mu.Lock()
		entry = r.entries[inv.key]
		r.mu.Unlock()
	}
	if entry == nil {
		return
	}
	r.mu.Lock()
	entry.refs += delta
	entry.lastAccess = time.Now()
	r.mu.Unlock()
}

func (r *invocationRegistry) touch(inv pluginInvocation) {
	entry := r.ensure(inv)
	if entry == nil {
		return
	}
	r.mu.Lock()
	entry.lastAccess = time.Now()
	r.mu.Unlock()
}

func (r *invocationRegistry) ensure(inv pluginInvocation) *invocationEntry {
	source := inv.plugin.WatchSource()
	if source == nil {
		return nil
	}

	r.mu.Lock()
	if entry, ok := r.entries[inv.key]; ok {
		r.mu.Unlock()
		return entry
	}
	if done, ok := r.pending[inv.key]; ok {
		r.mu.Unlock()
		<-done
		return r.ensure(inv)
	}
	done := make(chan struct{})
	r.pending[inv.key] = done
	r.mu.Unlock()
	defer close(done)

	cancel, err := source.Acquire(inv.key)

	r.mu.Lock()
	delete(r.pending, inv.key)
	var entry *invocationEntry
	if err == nil {
		entry = &invocationEntry{cancel: cancel, lastAccess: time.Now()}
		r.entries[inv.key] = entry
	}
	r.mu.Unlock()
	return entry
}

func (r *invocationRegistry) evictIdle(maxIdle time.Duration) {
	now := time.Now()
	r.evict(func(entry *invocationEntry) bool {
		return now.Sub(entry.lastAccess) > maxIdle
	})
}

func (r *invocationRegistry) close() {
	r.evict(func(*invocationEntry) bool {
		return true
	})
}

func (r *invocationRegistry) evict(shouldEvict func(*invocationEntry) bool) {
	r.mu.Lock()
	var toCancel []func()
	for key, entry := range r.entries {
		if entry.refs > 0 {
			continue
		}
		if !shouldEvict(entry) {
			continue
		}
		toCancel = append(toCancel, entry.cancel)
		delete(r.entries, key)
	}
	r.mu.Unlock()
	for _, cancel := range toCancel {
		if cancel != nil {
			cancel()
		}
	}
}
