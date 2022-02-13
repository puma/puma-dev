package linebuffer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineBuffer(t *testing.T) {
	t.Run("appends lines to the buffer", func(t *testing.T) {
		var lb LineBuffer

		lb.Append("hello")

		var buf bytes.Buffer

		n, err := lb.WriteTo(&buf)
		assert.NoError(t, err)

		assert.Equal(t, int64(5), n)
	})

	t.Run("wraps around automatically", func(t *testing.T) {
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

	t.Run("wraps around automatically multiple times", func(t *testing.T) {
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

}
