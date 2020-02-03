package main

import (
	"fmt"
	"os"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	EnsurePumaDevDirectory()
	os.Exit(m.Run())
}

func TestCommand_noCommandArg(t *testing.T) {
	StubFlagArgs(nil)
	err := command()
	assert.Equal(t, "Unknown command: \n", err.Error())
}

func TestCommand_badCommandArg(t *testing.T) {
	StubFlagArgs([]string{"doesnotexist"})
	err := command()
	assert.Equal(t, "Unknown command: doesnotexist\n", err.Error())
}

func TestCommand_link_noArgs(t *testing.T) {
	StubFlagArgs([]string{"link"})

	appDir, _ := homedir.Expand("~/my-test-puma-dev-application")

	WithWorkingDirectory(appDir, func() {
		actual := WithStdoutCaptured(func() {
			if err := command(); err != nil {
				assert.Fail(t, err.Error())
			}
		})

		expected := fmt.Sprintf("+ App 'my-test-puma-dev-application' created, linked to '%s'\n", appDir)
		assert.Equal(t, expected, actual)
	})

	RemoveAppSymlinkOrFail(t, "my-test-puma-dev-application")
}

func TestCommand_link_withNameOverride(t *testing.T) {
	tmpCwd := "/tmp/puma-dev-example-command-link-noargs"

	StubFlagArgs([]string{"link", "-n", "anothername", tmpCwd})

	WithWorkingDirectory(tmpCwd, func() {
		actual := WithStdoutCaptured(func() {
			if err := command(); err != nil {
				assert.Fail(t, err.Error())
			}
		})

		assert.Equal(t, "+ App 'anothername' created, linked to '/tmp/puma-dev-example-command-link-noargs'\n", actual)
	})

	RemoveAppSymlinkOrFail(t, "anothername")
}

func TestCommand_link_invalidDirectory(t *testing.T) {
	StubFlagArgs([]string{"link", "/this/path/does/not/exist"})

	err := command()

	assert.Equal(t, "Invalid directory: /this/path/does/not/exist", err.Error())
}

func TestCommand_link_reassignExistingApp(t *testing.T) {
	appAlias := "apptastic"
	appDir1 := MakeDirectoryOrFail(t, "/tmp/puma-dev-test-command-link-reassign-existing-app-one")
	appDir2 := MakeDirectoryOrFail(t, "/tmp/puma-dev-test-command-link-reassign-existing-app-two")

	defer RemoveDirectoryOrFail(t, appDir1)
	defer RemoveDirectoryOrFail(t, appDir2)
	defer RemoveAppSymlinkOrFail(t, appAlias)

	StubFlagArgs([]string{"link", "-n", appAlias, appDir1})
	actual1 := WithStdoutCaptured(func() {
		if err := command(); err != nil {
			assert.Fail(t, err.Error())
		}
	})

	assert.Equal(t, fmt.Sprintf("+ App '%s' created, linked to '%s'\n", appAlias, appDir1), actual1)

	StubFlagArgs([]string{"link", "-n", appAlias, appDir2})
	actual2 := WithStdoutCaptured(func() {
		if err := command(); err != nil {
			assert.Fail(t, err.Error())
		}
	})

	assert.Equal(t, fmt.Sprintf("! App '%s' already exists, pointed at '%s'\n", appAlias, appDir1), actual2)
}
