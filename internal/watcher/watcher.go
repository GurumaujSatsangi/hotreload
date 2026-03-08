package watcher

import "github.com/fsnotify/fsnotify"

type Watcher struct {
	watcher *fsnotify.Watcher
	events  chan string
}

func NewWatcher(root string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := w.Add(root); err != nil {
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
