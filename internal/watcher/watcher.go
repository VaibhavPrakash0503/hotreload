package watcher

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	root      string
	fsWatcher *fsnotify.Watcher
	filter    *Filter
}

func NewWatcher(root string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		root:      root,
		fsWatcher: fsWatcher,
		filter:    NewFilter(),
	}

	err = w.addDirectory(root)
	if err != nil {
		fsWatcher.Close()
		return nil, err
	}

	return w, nil
}

func (w *Watcher) addDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if w.filter.ShouldIgnoreDir(info.Name()) {
			return filepath.SkipDir
		}

		if err := w.fsWatcher.Add(path); err != nil {
			return err
		}

		slog.Debug("Watching directory", "path", path)

		return nil
	})
}

func (w *Watcher) Watch(events chan<- string) {
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			if event.Op == fsnotify.Chmod {
				continue
			}

			if w.filter.ShouldIgnoreExt(filepath.Ext(event.Name)) {
				continue
			}

			slog.Info(event.Op.String())
			slog.Debug("File changed", "event", event.Name)
			events <- event.Name

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			slog.Error("Watcher error", "error", err)
		}
	}
}

func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}
