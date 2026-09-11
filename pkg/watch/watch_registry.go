package watch

import (
	"sync"
	"time"
)

type watchEntry struct {
	input       watchConfig
	invocations []pluginInvocation
	subscribers map[sink]struct{}
	idleSince   time.Time

	lastDelivered    string
	hasLastDelivered bool
}

type watchRegistry struct {
	mu          sync.Mutex
	entries     map[WatchKey]*watchEntry
	invocations *invocationRegistry
}

func newWatchRegistry(invocations *invocationRegistry) *watchRegistry {
	return &watchRegistry{entries: map[WatchKey]*watchEntry{}, invocations: invocations}
}

func (r *watchRegistry) input(key WatchKey) (watchConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.entries[key]
	if !ok {
		return watchConfig{}, false
	}
	return entry.input, true
}

func (r *watchRegistry) ensure(key WatchKey, input watchConfig, s sink) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.entries[key]
	if !ok {
		r.entries[key] = &watchEntry{
			input:       input,
			subscribers: map[sink]struct{}{s: {}},
		}
		return
	}
	if len(entry.subscribers) == 0 {
		entry.invocations = nil
	}
	entry.subscribers[s] = struct{}{}
	entry.idleSince = time.Time{}
}

func (r *watchRegistry) update(key WatchKey, result Result, invocations []pluginInvocation) (subs []sink, changed bool) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok {
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
	changed = result.Err != nil || !entry.hasLastDelivered || entry.lastDelivered != result.Output
	if result.Err == nil {
		entry.lastDelivered = result.Output
		entry.hasLastDelivered = true
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
	}
	r.mu.Unlock()

	r.ref(deactivated, -1)
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
