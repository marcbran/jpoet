package watch

import (
	"context"
	"sync"
	"time"
)

type watchEntry struct {
	input       watchConfig
	invocations []pluginInvocation
	subscribers map[sink]struct{}
	idleSince   time.Time

	lastIdentity    string
	hasLastIdentity bool

	started uint64
	cancel  context.CancelFunc

	deliverMu sync.Mutex
	delivered uint64
}

type watchRegistry struct {
	mu          sync.Mutex
	entries     map[WatchKey]*watchEntry
	invocations *invocationRegistry
}

func newWatchRegistry(invocations *invocationRegistry) *watchRegistry {
	return &watchRegistry{entries: map[WatchKey]*watchEntry{}, invocations: invocations}
}

func (r *watchRegistry) ensure(key WatchKey, input watchConfig, s sink) (watchConfig, uint64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.entries[key]
	if !ok {
		entry = &watchEntry{
			input:       input,
			subscribers: map[sink]struct{}{},
		}
		r.entries[key] = entry
	}
	first := len(entry.subscribers) == 0
	if first {
		entry.invocations = nil
		entry.started++
	}
	entry.subscribers[s] = struct{}{}
	entry.idleSince = time.Time{}
	return entry.input, entry.started, first
}

func (r *watchRegistry) supersede(parent context.Context, key WatchKey) (watchConfig, context.Context, context.CancelFunc, uint64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.entries[key]
	if !ok {
		return watchConfig{}, nil, nil, 0, false
	}
	if entry.cancel != nil {
		entry.cancel()
	}
	ctx, cancel := context.WithCancel(parent)
	entry.cancel = cancel
	entry.started++
	return entry.input, ctx, cancel, entry.started, true
}

func (r *watchRegistry) cancelRunning() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range r.entries {
		if entry.cancel != nil {
			entry.cancel()
			entry.cancel = nil
		}
	}
}

func (r *watchRegistry) deliver(key WatchKey, seq uint64, result Result, subs []sink) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	r.mu.Unlock()
	if !ok {
		return
	}
	entry.deliverMu.Lock()
	defer entry.deliverMu.Unlock()
	if seq < entry.delivered {
		return
	}
	entry.delivered = seq
	for _, s := range subs {
		s.deliver(result)
	}
}

func (r *watchRegistry) update(key WatchKey, seq uint64, result Result, invocations []pluginInvocation) (subs []sink, changed bool) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok || entry.started != seq {
		r.mu.Unlock()
		return nil, false
	}
	var activated, deactivated []pluginInvocation
	if len(entry.subscribers) > 0 {
		for _, inv := range invocations {
			if !containsInvocation(entry.invocations, inv) {
				activated = append(activated, inv)
			}
		}
		for _, inv := range entry.invocations {
			if !containsInvocation(invocations, inv) {
				deactivated = append(deactivated, inv)
			}
		}
	}
	entry.invocations = invocations
	identity := identityOf(entry.input.identity, result.Output)
	changed = result.Err != nil || !entry.hasLastIdentity || entry.lastIdentity != identity
	if result.Err == nil {
		entry.lastIdentity = identity
		entry.hasLastIdentity = true
	}
	subs = make([]sink, 0, len(entry.subscribers))
	for s := range entry.subscribers {
		subs = append(subs, s)
	}
	r.mu.Unlock()

	r.ref(activated, 1)
	r.ref(deactivated, -1)
	return subs, changed
}

func (r *watchRegistry) detach(key WatchKey, s sink) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok {
		r.mu.Unlock()
		return
	}
	if _, present := entry.subscribers[s]; !present {
		r.mu.Unlock()
		return
	}
	delete(entry.subscribers, s)
	var deactivated []pluginInvocation
	if len(entry.subscribers) == 0 {
		entry.idleSince = time.Now()
		deactivated = entry.invocations
		if entry.cancel != nil {
			entry.cancel()
			entry.cancel = nil
		}
	}
	r.mu.Unlock()

	r.ref(deactivated, -1)
}

func (r *watchRegistry) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range r.entries {
		for s := range entry.subscribers {
			s.close()
		}
		entry.subscribers = map[sink]struct{}{}
	}
}

func (r *watchRegistry) ref(invocations []pluginInvocation, delta int) {
	for _, inv := range invocations {
		r.invocations.ref(inv, delta)
	}
}

func (r *watchRegistry) evictIdle(maxIdle time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for key, entry := range r.entries {
		if len(entry.subscribers) > 0 {
			continue
		}
		if entry.idleSince.IsZero() {
			continue
		}
		if now.Sub(entry.idleSince) <= maxIdle {
			continue
		}
		delete(r.entries, key)
	}
}

func (r *watchRegistry) matchingInvocation(change pluginInvocation) []WatchKey {
	r.mu.Lock()
	defer r.mu.Unlock()
	var keys []WatchKey
	for key, entry := range r.entries {
		if containsInvocation(entry.invocations, change) {
			keys = append(keys, key)
		}
	}
	return keys
}

func identityOf(identity WatchIdentity, output string) string {
	if identity == nil {
		return output
	}
	value, ok := identity(output)
	if !ok {
		return output
	}
	return value
}

func containsInvocation(invocations []pluginInvocation, target pluginInvocation) bool {
	for _, inv := range invocations {
		if inv.key == target.key && inv.plugin == target.plugin {
			return true
		}
	}
	return false
}

func (r *watchRegistry) subscribedKeys() []WatchKey {
	r.mu.Lock()
	defer r.mu.Unlock()
	var keys []WatchKey
	for key, entry := range r.entries {
		if len(entry.subscribers) > 0 {
			keys = append(keys, key)
		}
	}
	return keys
}

func (r *watchRegistry) hasSubscribers(key WatchKey) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.entries[key]
	if !ok {
		return false
	}
	return len(entry.subscribers) > 0
}
