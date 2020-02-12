package main

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
)

func TestMain_allCheck_versionFlag(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubFlagArgs([]string{"-V"})
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
		StubFlagArgs([]string{"-badarg"})
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
