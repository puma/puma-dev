package dev

import (
	"os"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/stretchr/testify/assert"
)

var testHttp HTTPServer

func TestMain(m *testing.M) {
	EnsurePumaDevDirectory()
	os.Exit(m.Run())
}

func TestHttp_removeTLD_test(t *testing.T) {
	str := testHttp.removeTLD("psychic-octo-giggle.test")

	assert.Equal(t, "psychic-octo-giggle", str)
}

func TestHttp_removeTLD_xipIoDots(t *testing.T) {
	str := testHttp.removeTLD("legendary-meme.0.0.0.0.xip.io")

	assert.Equal(t, "legendary-meme", str)
}

func TestHttp_removeTLD_nipIoDots(t *testing.T) {
	str := testHttp.removeTLD("effective-invention.255.255.255.255.nip.io")

	assert.Equal(t, "effective-invention", str)
}
