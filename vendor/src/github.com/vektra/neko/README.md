neko - a simple golang test organizer
=====================================

[Doc's from Godoc.org!](http://godoc.org/github.com/vektra/neko)

There are many tools for improving Go's tests.

It's hard to beat the simplicity of `go test` but we all know it can get, well,
a little disorganized.

`neko` helps by just give you an extra little bit of organization to perform
common setup between tests.

Oh, and it integrates with `github.com/stretchr/testify/mock` to coordinate
your mocks (clearing and asserting them).

Here is a quick example:

```go
import (
  "testing"
  "github.com/vektra/neko"
)

func TestNekoEnjoysFun(t *testing.T) {
	n := neko.Start(t)

	var fun Fun

	n.Setup(func() {
		fun = CreateAmeowsements()
	})

	n.It("enjoys fun", func() {
		if !fun.IsFun() {
			t.Fatal("fun isn't fun?? :( :(")
		}
	})

	n.It("knows when it's fun time", func() {
		if !fun.ItsTime() {
			t.Fatal("no fun time? :( :(")
		}
	})

  n.Meow()
}
```
