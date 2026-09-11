package watch

import (
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type directoryWatcher struct {
	fsWatcher *fsnotify.Watcher

	mu           sync.Mutex
	watchedDirs  map[string]struct{}
	trackedFiles map[string]struct{}
}

func newDirectoryWatcher() (*directoryWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &directoryWatcher{
		fsWatcher:    fsWatcher,
		watchedDirs:  map[string]struct{}{},
		trackedFiles: map[string]struct{}{},
	}, nil
}

func (w *directoryWatcher) track(paths []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range paths {
		w.trackedFiles[p] = struct{}{}
		dir := filepath.Dir(p)
		if _, ok := w.watchedDirs[dir]; ok {
			continue
		}
		if err := w.fsWatcher.Add(dir); err == nil {
			w.watchedDirs[dir] = struct{}{}
		}
	}
}

const contentOps = fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename

func (w *directoryWatcher) run(done <-chan struct{}, onChange func(path string)) {
	for {
		select {
		case <-done:
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			if event.Op&contentOps == 0 {
				continue
			}
			w.mu.Lock()
			_, tracked := w.trackedFiles[event.Name]
			w.mu.Unlock()
			if !tracked {
				continue
			}
			onChange(event.Name)
		case _, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *directoryWatcher) Close() error {
	return w.fsWatcher.Close()
}
