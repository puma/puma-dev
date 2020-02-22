package devtest

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fTribe = flag.Bool("cankickit", false, "Can I kick it?")

func TestStubCommandLineArgs(t *testing.T) {
	StubCommandLineArgs()
	assert.False(t, *fTribe)
	StubCommandLineArgs("-cankickit")
	assert.True(t, *fTribe)
	StubCommandLineArgs("-cankickit=false")
	assert.False(t, *fTribe)
	StubCommandLineArgs("-cankickit=true")
	assert.True(t, *fTribe)
	StubCommandLineArgs()
	assert.False(t, *fTribe)
	StubCommandLineArgs("-cankickit")
	assert.True(t, *fTribe)
	StubCommandLineArgs()
	assert.False(t, *fTribe)
}
