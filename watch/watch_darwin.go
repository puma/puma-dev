package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsevents"
)

func Watch(watchedPath string, done <-chan struct{}, change func()) error {
	fmt.Printf("watchedPath: %s\n", watchedPath)
	watchedAbsPath, err := filepath.EvalSymlinks(watchedPath)
	if err != nil {
		return err
	}
	fmt.Printf("watchedAbsPath: %s\n", watchedAbsPath)

	lastStat, err := os.Stat(watchedAbsPath)
	if err != nil {
		return err
	}

	dev, err := fsevents.DeviceForPath(watchedAbsPath)
	if err != nil {
		return err
	}

	es := &fsevents.EventStream{
		Paths:   []string{watchedAbsPath},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.IgnoreSelf,
	}

	es.Start()

	defer es.Stop()

	for {
		select {
		case <-es.Events:
			cur, err := os.Stat(watchedAbsPath)
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
