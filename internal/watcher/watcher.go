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

// NewWatcher initializes a new Watcher for the specified root directory
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

// Add Directory and its subdirectories to the watcher, skipping ignored directories
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

// Watch for file system events and send changed file paths to the events channel
func (w *Watcher) Watch(events chan<- string) {
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			// If a new directory was created, watch it and its subtree.
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if !w.filter.ShouldIgnoreDir(info.Name()) {
						if err := w.addDirectory(event.Name); err != nil {
							slog.Warn("Failed to watch new directory", "path", event.Name, "error", err)
						} else {
							slog.Debug("Now watching new directory", "path", event.Name)
						}
					}
					// Don't forward bare dir-create as a reload trigger.
					continue
				}
			}

			if w.filter.ShouldIgnoreExt(event.Name) {
				continue
			}

			slog.Debug("File changed", "op", event.Op.String(), "event", event.Name)
			events <- event.Name

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			slog.Warn("Watcher error", "error", err)
		}
	}
}

func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}
