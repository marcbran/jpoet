package watch

import (
	"os"
	"path/filepath"
)

type sink interface {
	deliver(Result)
}

type channelSink struct {
	ch chan Result
}

func (s channelSink) deliver(r Result) {
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
