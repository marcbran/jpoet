package jpoet

import (
	"context"
	"sync"
)

type Lifecycle struct {
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	cleanup func() error

	closedMu  sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func NewLifecycle(cleanup func() error) *Lifecycle {
	ctx, cancel := context.WithCancel(context.Background())
	return &Lifecycle{ctx: ctx, cancel: cancel, cleanup: cleanup}
}

func (l *Lifecycle) Enter() error {
	l.closedMu.Lock()
	defer l.closedMu.Unlock()
	if l.closed {
		return ErrEnvironmentClosed
	}
	l.wg.Add(1)
	return nil
}

func (l *Lifecycle) Leave() {
	l.wg.Done()
}

func (l *Lifecycle) Go(f func()) {
	l.wg.Go(f)
}

func (l *Lifecycle) Done() <-chan struct{} {
	return l.ctx.Done()
}

func (l *Lifecycle) Close() error {
	var err error
	l.closeOnce.Do(func() {
		l.closedMu.Lock()
		l.closed = true
		l.closedMu.Unlock()

		l.cancel()
		l.wg.Wait()

		if l.cleanup != nil {
			err = l.cleanup()
		}
	})
	return err
}
