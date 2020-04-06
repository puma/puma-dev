package watch

import (
	"os"

	"github.com/fsnotify/fsnotify"
)

func Watch(restart string, done <-chan struct{}, change func()) error {
	lastStat, err := os.Stat(restart)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = watcher.Add(restart)
	if err != nil {
		return err
	}

	defer watcher.Close()

	for {
		select {
		case <-watcher.Events:
			cur, err := os.Stat(restart)
			if err != nil {
				return err
			}

			if cur.ModTime().After(lastStat.ModTime()) {
				change()
			}
		case <-done:
			return nil
		}
	}
}
