package watch

import (
	"sync"
	"time"
)

type watchEntry struct {
	input       watchConfig
	invocations []pluginInvocation
	lastOutput  string
	subscribers map[chan string]struct{}
	idleSince   time.Time
}

type watchRegistry struct {
	mu          sync.Mutex
	entries     map[WatchKey]*watchEntry
	invocations *invocationRegistry
}

func newWatchRegistry(invocations *invocationRegistry) *watchRegistry {
	return &watchRegistry{entries: map[WatchKey]*watchEntry{}, invocations: invocations}
}

func (r *watchRegistry) ref(invocations []pluginInvocation, delta int) {
	for _, inv := range invocations {
		r.invocations.ref(inv, delta)
	}
}

func (r *watchRegistry) matching(change pluginInvocation) []WatchKey {
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

func (r *watchRegistry) input(key WatchKey) (watchConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.entries[key]
	if !ok {
		return watchConfig{}, false
	}
	return entry.input, true
}

func (r *watchRegistry) attach(key WatchKey) (initial string, ch chan string, ok bool) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok {
		r.mu.Unlock()
		return "", nil, false
	}
	wasIdle := len(entry.subscribers) == 0
	ch = make(chan string, 1)
	entry.subscribers[ch] = struct{}{}
	entry.idleSince = time.Time{}
	var activated []pluginInvocation
	if wasIdle {
		activated = entry.invocations
	}
	initial = entry.lastOutput
	r.mu.Unlock()

	r.ref(activated, 1)
	return initial, ch, true
}

func (r *watchRegistry) create(key WatchKey, input watchConfig, output string, invocations []pluginInvocation) (initial string, ch chan string) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok {
		ch = make(chan string, 1)
		r.entries[key] = &watchEntry{
			input:       input,
			invocations: invocations,
			lastOutput:  output,
			subscribers: map[chan string]struct{}{ch: {}},
		}
		r.mu.Unlock()

		r.ref(invocations, 1)
		return output, ch
	}
	wasIdle := len(entry.subscribers) == 0
	ch = make(chan string, 1)
	entry.subscribers[ch] = struct{}{}
	entry.idleSince = time.Time{}
	var activated []pluginInvocation
	if wasIdle {
		activated = entry.invocations
	}
	initial = entry.lastOutput
	r.mu.Unlock()

	r.ref(activated, 1)
	return initial, ch
}

func (r *watchRegistry) update(key WatchKey, output string, invocations []pluginInvocation) (subs []chan string) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok {
		r.mu.Unlock()
		return nil
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
	entry.lastOutput = output
	subs = make([]chan string, 0, len(entry.subscribers))
	for ch := range entry.subscribers {
		subs = append(subs, ch)
	}
	r.mu.Unlock()

	r.ref(activated, 1)
	r.ref(deactivated, -1)
	return subs
}

func (r *watchRegistry) detach(key WatchKey, ch chan string) {
	r.mu.Lock()
	entry, ok := r.entries[key]
	if !ok {
		r.mu.Unlock()
		return
	}
	if _, present := entry.subscribers[ch]; !present {
		r.mu.Unlock()
		return
	}
	delete(entry.subscribers, ch)
	var deactivated []pluginInvocation
	if len(entry.subscribers) == 0 {
		entry.idleSince = time.Now()
		deactivated = entry.invocations
	}
	r.mu.Unlock()

	r.ref(deactivated, -1)
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

func containsInvocation(invocations []pluginInvocation, target pluginInvocation) bool {
	for _, inv := range invocations {
		if inv.key == target.key && inv.plugin == target.plugin {
			return true
		}
	}
	return false
}
