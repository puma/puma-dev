package linebuffer

import (
	"io"
	"sync"
)

const DefaultSize = 1024

type LineBuffer struct {
	Size int

	lock  sync.Mutex
	cur   int
	lines []string
}

func (lb *LineBuffer) Append(line string) error {
	lb.lock.Lock()
	defer lb.lock.Unlock()

	if lb.Size == 0 {
		lb.Size = DefaultSize
	}

	if len(lb.lines) < lb.Size {
		lb.lines = append(lb.lines, line)
	} else {
		lb.lines[lb.cur] = line
		lb.cur++

		if lb.cur == len(lb.lines) {
			lb.cur = 0
		}
	}

	return nil
}

func (lb *LineBuffer) Do(x func(string) error) error {
	lb.lock.Lock()
	defer lb.lock.Unlock()

	var err error

	if len(lb.lines) < lb.Size {
		for _, l := range lb.lines {
			err = x(l)
			if err != nil {
				return err
			}
		}

		return nil
	}

	for i := lb.cur; i < lb.Size; i++ {
		err = x(lb.lines[i])
		if err != nil {
			return err
		}
	}

	for i := 0; i < lb.cur; i++ {
		err = x(lb.lines[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (lb *LineBuffer) WriteTo(w io.Writer) (int64, error) {
	var tot int64

	err := lb.Do(func(l string) error {
		n, err := w.Write([]byte(l))
		if err != nil {
			return err
		}
		tot += int64(n)
		return nil
	})

	return tot, err
}
