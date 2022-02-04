package dev

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPumaCommand_withEnv(t *testing.T) {
	t.Setenv("CONFIG", "-")
	t.Setenv("WORKERS", "4")
	t.Setenv("THREADS", "5")

	cmd, _ := BuildPumaCommand("foo", "/tmp/tmp/foo.sock", "/tmp")

	assert.Regexp(t, " -w4 ", cmd.Args[4])
	assert.Regexp(t, " -t0:5 ", cmd.Args[4])
	assert.Regexp(t, " -C- ", cmd.Args[4])
}

func TestBuildPumaCommand_withoutEnv(t *testing.T) {
	cmd, _ := BuildPumaCommand("foo", "/tmp/tmp/foo.sock", "/tmp")

	assert.NotRegexp(t, " -w", cmd.Args[4])
	assert.NotRegexp(t, " -t", cmd.Args[4])
	assert.NotRegexp(t, " -C", cmd.Args[4])
}
