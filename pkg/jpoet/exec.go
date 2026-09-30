package jpoet

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrNoSuchPlugin = errors.New("no such plugin")
	ErrNoSuchAction = errors.New("no such action")
)

func (e *Environment) Exec(ctx context.Context, plugin string, action string, data map[string]any) (string, error) {
	if err := e.lifecycle.Enter(); err != nil {
		return "", err
	}
	defer e.lifecycle.Leave()

	fn, err := e.findAction(plugin, action)
	if err != nil {
		return "", err
	}
	return fn(ctx, data)
}

func (e *Environment) findAction(plugin string, action string) (ActionFunc, error) {
	var plugins []string
	for _, p := range e.plugins {
		plugins = append(plugins, p.Name())
		if p.Name() != plugin {
			continue
		}
		fn, ok := p.Action(action)
		if !ok {
			return nil, fmt.Errorf("%w: %s, available actions: %s", ErrNoSuchAction, action, strings.Join(p.ActionNames(), ", "))
		}
		return fn, nil
	}
	sort.Strings(plugins)
	return nil, fmt.Errorf("%w: %s, available plugins: %s", ErrNoSuchPlugin, plugin, strings.Join(plugins, ", "))
}
