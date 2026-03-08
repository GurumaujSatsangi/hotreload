package watcher

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
	events  chan string
}

func NewWatcher(root string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := addDirectoryTree(w, root); err != nil {
		_ = w.Close()
		return nil, err
	}

	instance := &Watcher{
		watcher: w,
		events:  make(chan string),
	}

	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					close(instance.events)
					return
				}

				if event.Op&fsnotify.Create != 0 {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						_ = addDirectoryTree(w, event.Name)
					}
				}

				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					if _, err := os.Stat(event.Name); err != nil {
						_ = w.Remove(event.Name)
					}
				}

				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
					instance.events <- event.Name
				}
			case _, ok := <-w.Errors:
				if !ok {
					close(instance.events)
					return
				}
			}
		}
	}()

	return instance, nil
}

func (w *Watcher) Events() <-chan string {
	return w.events
}

func addDirectoryTree(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if addErr := w.Add(path); addErr != nil {
			if os.IsNotExist(addErr) {
				return nil
			}
			return addErr
		}
		return nil
	})
}
