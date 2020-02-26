package devtest

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fTribe = flag.Bool("cankickit", false, "Can I kick it?")

func TestStubCommandLineArgs_flagReset(t *testing.T) {
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

func TestStubCommandLineArgs_argReset(t *testing.T) {
	StubCommandLineArgs("one", "two")
	assert.Equal(t, 2, flag.NArg())
	assert.Equal(t, []string{"one", "two"}, flag.Args())
	StubCommandLineArgs()
	assert.Equal(t, 0, flag.NArg())
	assert.Equal(t, []string{}, flag.Args())
}
