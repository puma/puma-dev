package linebuffer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestLineBuffer(t *testing.T) {
	n := neko.Start(t)

	n.It("appends lines to the buffer", func() {
		var lb LineBuffer

		lb.Append("hello")

		var buf bytes.Buffer

		n, err := lb.WriteTo(&buf)
		require.NoError(t, err)

		assert.Equal(t, int64(5), n)
	})

	n.It("wraps around automatically", func() {
		var lb LineBuffer

		lb.Size = 3

		lb.Append("hello1")
		lb.Append("hello2")
		lb.Append("hello3")
		lb.Append("hello4")

		var lines []string

		lb.Do(func(x string) error {
			lines = append(lines, x)
			return nil
		})

		assert.Equal(t, 3, len(lb.lines))

		assert.Equal(t, "hello2", lines[0])
		assert.Equal(t, "hello3", lines[1])
		assert.Equal(t, "hello4", lines[2])
	})

	n.It("wraps around automatically multiple times", func() {
		var lb LineBuffer

		lb.Size = 3

		lb.Append("hello1")
		lb.Append("hello2")
		lb.Append("hello3")
		lb.Append("hello4")
		lb.Append("hello5")
		lb.Append("hello6")
		lb.Append("hello7")

		var lines []string

		lb.Do(func(x string) error {
			lines = append(lines, x)
			return nil
		})

		assert.Equal(t, 3, len(lb.lines))

		assert.Equal(t, "hello5", lines[0])
		assert.Equal(t, "hello6", lines[1])
		assert.Equal(t, "hello7", lines[2])
	})

	n.Meow()
}
