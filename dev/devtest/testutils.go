package devtest

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

var (
	appSymlinkHome      = "~/.puma-dev"
	DebugLoggingEnabled = os.Getenv("DEBUG_LOG") == "1"
	StubbedArgs         = make(map[string]int)

	_, b, _, _  = runtime.Caller(0)
	ProjectRoot = filepath.Join(filepath.Dir(b), "..", "..")
)

// SetFlagOrFail sets the value of a flag or fails the given testing.T
func SetFlagOrFail(t *testing.T, flagName string, flagValue string) {
	if err := flag.Set(flagName, flagValue); err != nil {
		assert.Fail(t, err.Error())
	}
}

// LogDebugf prints a formatted log message if DEBUG_LOG=1
func LogDebugf(msg string, vars ...interface{}) {
	if DebugLoggingEnabled {
		log.Printf(strings.Join([]string{"[DEBUG]", msg}, " "), vars...)
	}
}

/*
	StubCommandLineArgs overrides command arguments to allow flag-based branches
	to execute. It does not modify os.Args[0] so it can be used for subprocess
	tests. It also resets all defined flags to their default values, as
	`flag.Parse()` will not reset non-existent boolean flags if they have been
	stubbed in other tests.
*/
func StubCommandLineArgs(args ...string) {
	for _, arg := range args {
		StubbedArgs[arg] += 1
	}

	for arg := range StubbedArgs {
		stubbedFlagName := strings.Replace(arg, "-", "", -1)

		if fl := flag.Lookup(stubbedFlagName); fl != nil {

			oldValue := fl.Value.String()

			if err := flag.Set(fl.Name, fl.DefValue); err == nil {
				LogDebugf("reset flag %s to %s (was %s)", fl.Name, fl.Value.String(), oldValue)
			}
		}
	}

	os.Args = append([]string{os.Args[0]}, args...)
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

// WithStdoutCaptured executes the passed function and returns a string
// containing the stdout of the executed function.
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
	if !DirExists(path) {
		assert.FailNow(t, fmt.Sprintf("%s does not exist", path))
	}

	if err := os.RemoveAll(path); err != nil {
		assert.Fail(t, err.Error())
	} else {
		LogDebugf("removed %s\n", path)
	}
}

// MakeDirectoryOrFail makes a directory or fails the test, returning the path
// of the directory that was created.
func MakeDirectoryOrFail(t *testing.T, path string) func() {
	if err := os.MkdirAll(path, 0755); err != nil {
		assert.Fail(t, err.Error())
	}

	return func() {
		RemoveDirectoryOrFail(t, path)
	}
}

// WithWorkingDirectory executes the passed function within the context of
// the passed working directory path. If the directory does not exist, panic.
func WithWorkingDirectory(path string, f func()) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(err)
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

// FileExists returns true if a regular file exists at the given path.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists returns true if a directory exists at the given path.
func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
