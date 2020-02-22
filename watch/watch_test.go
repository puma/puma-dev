package watch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	devtest "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

var (
	tmpDir        = filepath.Join(devtest.ProjectRoot, "tmp")
	tmpFilename   = "restart.txt"
	tmpRestartTxt = filepath.Join(tmpDir, tmpFilename)
)

type Notice struct{}

func TestWatch_ExpectTimeout(t *testing.T) {
	defer createTmpDir(t)()
	touchFile(t, tmpRestartTxt)

	watchTriggered := watchTmpFileWithTimeout(t, func() {})

	assert.False(t, watchTriggered)
}

func TestWatch_ExpectTouchSignalAfterModify(t *testing.T) {
	defer createTmpDir(t)()
	touchFile(t, tmpRestartTxt)

	watchTriggered := watchTmpFileWithTimeout(t, func() {
		// HFS only has seconds resolution. We need to ensure that when we "touch"
		// the file, we advance the modified time by at least one second.
		time.Sleep(time.Second)
		touchFile(t, tmpRestartTxt)
	})

	assert.True(t, watchTriggered)
}

func createTmpDir(t *testing.T) func() {
	devtest.MakeDirectoryOrFail(t, tmpDir)

	return func() {
		devtest.RemoveDirectoryOrFail(t, tmpDir)
	}
}

func touchFile(t *testing.T, fullPath string) {
	if err := exec.Command("sh", "-c", fmt.Sprintf("touch %s", fullPath)).Run(); err != nil {
		assert.Fail(t, err.Error())
	}
}

func watchTmpFileWithTimeout(t *testing.T, f func()) bool {
	watchDone := make(chan struct{})
	watchTriggered := false

	if !devtest.FileExists(tmpRestartTxt) {
		assert.Fail(t, fmt.Sprintf("%s does not exist", tmpRestartTxt))
		return false
	}

	go func() {
		err := Watch(tmpRestartTxt, watchDone, func() {
			watchTriggered = true
			watchDone <- Notice{}
		})

		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				panic(err)
			}
		}
	}()

	timeoutDone := make(chan struct{})
	go func() {
		time.Sleep(2 * time.Second)
		timeoutDone <- Notice{}
	}()

	f()

	for {
		select {
		case <-watchDone:
			return watchTriggered
		case <-timeoutDone:
			return false
		}
	}
}
