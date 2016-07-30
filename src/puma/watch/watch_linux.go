package watch

import (
	"os"

	"golang.org/x/exp/inotify"
)

func Watch(restart string, done <-chan struct{}, change func()) error {
	lastStat, err := os.Stat(restart)
	if err != nil {
		return err
	}

	watcher, err := inotify.NewWatcher()
	if err != nil {
		return err
	}

	err = watcher.AddWatch(restart, inotify.IN_ATTRIB|inotify.IN_MODIFY)
	if err != nil {
		return err
	}

	defer watcher.Close()

	for {
		select {
		case <-watcher.Event:
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
