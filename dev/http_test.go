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
	str := testHttp.removeTLD("psychic-octo-guide.test")

	assert.Equal(t, "psychic-octo-guide", str)
}

func TestHttp_removeTLD_noTld(t *testing.T) {
	str := testHttp.removeTLD("shiny-train")

	assert.Equal(t, "shiny-train", str)
}

func TestHttp_removeTLD_mutlipartDomain(t *testing.T) {
	str := testHttp.removeTLD("expert-eureka.loc.al")

	assert.Equal(t, "expert-eureka.loc", str)
}

func TestHttp_removeTLD_dev(t *testing.T) {
	str := testHttp.removeTLD("bookish-giggle.dev:8080")

	assert.Equal(t, "bookish-giggle", str)
}

func TestHttp_removeTLD_xipIoMalformed(t *testing.T) {
	str := testHttp.removeTLD("legendary-meme.0.0.xip.io")

	assert.Equal(t, "", str)
}

func TestHttp_removeTLD_xipIoDots(t *testing.T) {
	str := testHttp.removeTLD("legendary-meme.0.0.0.0.xip.io")

	assert.Equal(t, "legendary-meme", str)
}

func TestHttp_removeTLD_nipIoDots(t *testing.T) {
	str := testHttp.removeTLD("effective-invention.255.255.255.255.nip.io")

	assert.Equal(t, "effective-invention", str)
}
