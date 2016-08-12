package dev

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/puma/puma-dev/linebuffer"
)

type Events struct {
	events linebuffer.LineBuffer
}

func (e *Events) Add(name string, args ...interface{}) string {
	var buf bytes.Buffer

	buf.WriteString("{")

	fmt.Fprintf(&buf, `"time":"%s","event":"%s"`, time.Now(), name)

	for i := 0; i < len(args); i += 2 {
		k := args[i]
		v := args[i+1]
		fmt.Fprintf(&buf, `,"%s":%#v`, k, v)
	}

	buf.WriteString("}\n")

	str := buf.String()

	e.events.Append(str)

	return str
}

func (e *Events) WriteTo(w io.Writer) (int64, error) {
	return e.events.WriteTo(w)
}
