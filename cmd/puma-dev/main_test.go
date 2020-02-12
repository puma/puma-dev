package main

import (
	"flag"
	"os"
	"os/exec"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

//
// func FlagResetForTesting() {
// 	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
// 	flag.CommandLine.Usage = flag.Usage()
// }

func TestMain_execWithExitStatus_versionFlag(t *testing.T) {
	StubCommandLineArgs([]string{"-V"})
	if err := flag.Set("V", "true"); err != nil {
		panic(err)
	}

	assert.True(t, *fVersion)

	execStdOut := WithStdoutCaptured(func() {
		status, shouldExit := execWithExitStatus()
		assert.Equal(t, 0, status)
		assert.Equal(t, true, shouldExit)
	})

	assert.Regexp(t, "^Version: devel \\(go[0-9.]+\\)\\n$", execStdOut)
}

func TestMain_execWithExitStatus_noFlag(t *testing.T) {
	StubCommandLineArgs(nil)
	if err := flag.Set("V", "false"); err != nil {
		panic(err)
	}

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
