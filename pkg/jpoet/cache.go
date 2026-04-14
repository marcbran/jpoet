package jpoet

import (
	"encoding/json"
	"sync"
	"time"
)

type cacheKey struct {
	funcName string
	args     string
}

type cacheEntry struct {
	value     any
	expiresAt time.Time
}

type Cache struct {
	mu   sync.Mutex
	data map[cacheKey]cacheEntry
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[cacheKey]cacheEntry),
	}
}

func (c *Cache) keyFor(funcName string, args []any) (cacheKey, bool) {
	b, err := json.Marshal(args)
	if err != nil {
		return cacheKey{}, false
	}
	return cacheKey{funcName: funcName, args: string(b)}, true
}

func (c *Cache) get(funcName string, args []any) (any, bool) {
	k, ok := c.keyFor(funcName, args)
	if !ok {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.data[k]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(c.data, k)
		return nil, false
	}
	return entry.value, true
}

func (c *Cache) store(funcName string, args []any, value any, ttl time.Duration) {
	if ttl == 0 {
		return
	}
	k, ok := c.keyFor(funcName, args)
	if !ok {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = cacheEntry{value: value, expiresAt: time.Now().Add(ttl)}
}

func (c *Cache) Writer(ttl func(funcName string, args []any) time.Duration) Middleware {
	return HookMiddleware(func(next Invoker, funcName string, args []any) (any, error) {
		result, err := next.Invoke(funcName, args)
		if err == nil {
			c.store(funcName, args, result, ttl(funcName, args))
		}
		return result, err
	})
}

func (c *Cache) Reader() Middleware {
	return HookMiddleware(func(next Invoker, funcName string, args []any) (any, error) {
		if v, ok := c.get(funcName, args); ok {
			return v, nil
		}
		return next.Invoke(funcName, args)
	})
}
