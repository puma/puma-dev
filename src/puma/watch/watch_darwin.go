package watch

import (
	"os"
	"time"

	"github.com/fsnotify/fsevents"
)

func Watch(restart string, done <-chan struct{}, change func()) error {
	lastStat, err := os.Stat(restart)
	if err != nil {
		return err
	}

	dev, err := fsevents.DeviceForPath(restart)
	if err != nil {
		return err
	}

	es := &fsevents.EventStream{
		Paths:   []string{restart},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.IgnoreSelf,
	}

	es.Start()

	defer es.Stop()

	for {
		select {
		case <-es.Events:
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
