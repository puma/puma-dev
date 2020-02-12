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
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestMain_allCheck_badArg(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubFlagArgs([]string{"bad-argument-that-does-not-exist"})
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_versionFlagWithArgs")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && e.Success() {
		t.Fatalf("process ran with err %v, want exit status 1", err)
	}
}
