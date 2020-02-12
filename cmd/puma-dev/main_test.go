package main

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

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

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		t.Fatalf("`puma-dev -V` had exit status %v, wanted exit status 0", err)
	}
}

func TestMain_allCheck_badArg(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubCommandLineArgs([]string{"-badarg"})
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_versionFlagWithArgs")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && e.Success() {
		t.Fatalf("`puma-dev -badarg` had exit status %v, wanted exit status 1", err)
	}
}
