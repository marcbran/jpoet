package watch

import (
	"os"
	"path/filepath"
	"sync"
)

type sink interface {
	deliver(Result)
	close()
}

type channelSink struct {
	ch     chan Result
	mu     sync.Mutex
	closed bool
}

func (s *channelSink) deliver(r Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.ch <- r:
	default:
		select {
		case <-s.ch:
		default:
		}
		select {
		case s.ch <- r:
		default:
		}
	}
}

func (s *channelSink) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.ch)
}

type fileSink struct {
	path string
}

func (s fileSink) deliver(r Result) {
	if r.Err != nil {
		return
	}
	existing, err := os.ReadFile(s.path)
	if err == nil && string(existing) == r.Output {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".tmp-*")
	if err != nil {
		return
	}
	_, err = tmp.WriteString(r.Output)
	closeErr := tmp.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(tmp.Name())
		return
	}
	_ = os.Rename(tmp.Name(), s.path)
}

func (s fileSink) close() {}
