package main

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	EnsurePumaDevDirectory()
	os.Exit(m.Run())
}

func TestMain_execWithExitStatus_versionFlag(t *testing.T) {
	StubCommandLineArgs([]string{"-V"})
	assert.True(t, *fVersion)

	execStdOut := WithStdoutCaptured(func() {
		status, shouldExit := execWithExitStatus()
		assert.Equal(t, 0, status)
		assert.Equal(t, true, shouldExit)
	})

	assert.Regexp(t, "^Version: devel \\(.+\\)\\n$", execStdOut)
}

func TestMain_execWithExitStatus_noFlag(t *testing.T) {
	StubCommandLineArgs(nil)
	assert.False(t, *fVersion)

	execStdOut := WithStdoutCaptured(func() {
		_, shouldExit := execWithExitStatus()
		assert.Equal(t, false, shouldExit)
	})

	assert.Equal(t, "", execStdOut)
}

func TestMain_allCheck_versionFlag(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubCommandLineArgs([]string{"-V"})
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_versionFlag")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	exit, ok := err.(*exec.ExitError)

	assert.False(t, ok)
	assert.Nil(t, exit)
}

func TestMain_allCheck_badArg(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubCommandLineArgs([]string{"-badarg"})
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_badArg")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	exit, ok := err.(*exec.ExitError)

	assert.True(t, ok)
	assert.False(t, exit.Success())
}
