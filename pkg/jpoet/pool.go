package jpoet

import (
	"context"
	"sync"

	"github.com/google/go-jsonnet"
)

type pooledVM struct {
	vm         *jsonnet.VM
	slot       *evalSlot
	importer   jsonnet.Importer
	generation uint64
}

func (v *pooledVM) refresh(generation uint64) {
	if v.generation == generation {
		return
	}
	if v.importer != nil {
		v.vm.Importer(v.importer)
	}
	v.generation = generation
}

type vmPool struct {
	build func() *pooledVM
	slots chan struct{}

	mu   sync.Mutex
	idle []*pooledVM
}

func newVMPool(size int, build func() *pooledVM) *vmPool {
	if size < 1 {
		size = 1
	}
	return &vmPool{build: build, slots: make(chan struct{}, size)}
}

func (p *vmPool) acquire(ctx context.Context) (*pooledVM, error) {
	select {
	case p.slots <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	n := len(p.idle)
	if n == 0 {
		return p.build(), nil
	}
	v := p.idle[n-1]
	p.idle = p.idle[:n-1]
	return v, nil
}

func (p *vmPool) release(v *pooledVM) {
	p.mu.Lock()
	p.idle = append(p.idle, v)
	p.mu.Unlock()
	<-p.slots
}
