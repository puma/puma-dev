package devtest

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

var (
	appSymlinkHome = "~/.puma-dev"
)

// StubFlagArgs overrides command arguments to pretend as if puma-dev was executed at the commandline.
// ex: StubArgFlags([]string{"-n", "myapp", "path/to/app"}) ->
//   $ puma-dev -n myapp path/to/app
func StubFlagArgs(args []string) {
	os.Args = append([]string{"puma-dev"}, args...)
	flag.Parse()
}

// EnsurePumaDevDirectory creates ~/.puma-dev if it does not already exist.
func EnsurePumaDevDirectory() {
	path, err := homedir.Expand(appSymlinkHome)

	if err != nil {
		panic(err)
	}

	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return
	}

	if err := os.Mkdir(path, 0755); err != nil {
		panic(err)
	}
}

// WithStdoutCaptured executes the passed function and returns a string containing the stdout of the executed function.
func WithStdoutCaptured(f func()) string {
	osStdout := os.Stdout
	r, w, err := os.Pipe()

	if err != nil {
		panic(err)
	}

	os.Stdout = w

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		if _, err := io.Copy(&buf, r); err != nil {
			panic(err)
		}

		outC <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = osStdout
	out := <-outC

	return out
}

// RemoveDirectoryOrFail removes a directory (and its contents) or fails the test.
func RemoveDirectoryOrFail(t *testing.T, path string) {
	if err := os.RemoveAll(path); err != nil {
		assert.Fail(t, err.Error())
	}
}

// MakeDirectoryOrFail makes a directory or fails the test, returning the path of the directory that was created.
func MakeDirectoryOrFail(t *testing.T, path string) string {
	if err := os.Mkdir(path, 0755); err != nil {
		assert.Fail(t, err.Error())
	}
	return path
}

// WithWorkingDirectory executes the passed function within the context of
// the passed working directory path. If the directory does not exist,
// WithWorkingDirectory will attempt to create it.
func WithWorkingDirectory(path string, f func()) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic(err)
		}
	}

	originalPath, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	if err := os.Chdir(path); err != nil {
		panic(err)
	}

	f()

	if err := os.Chdir(originalPath); err != nil {
		panic(err)
	}
}

// RemoveAppSymlinkOrFail deletes a symlink at ~/.puma-dev/{name} or fails the test.
func RemoveAppSymlinkOrFail(t *testing.T, name string) {
	path, err := homedir.Expand(filepath.Join(appSymlinkHome, name))
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	if err := os.Remove(path); err != nil {
		panic(err)
	}
}
