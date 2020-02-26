package dev

import (
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

func TestTrustCert_Linux_noCertProvided(t *testing.T) {
	stdOut := WithStdoutCaptured(func() {
		// TrustCert on Linux is a no-op, so it "succeeds" even if the path isn't found
		err := TrustCert("/does/not/exist")
		assert.Nil(t, err)
	})

	assert.Regexp(t, "^! Add /does/not/exist to your browser to trust CA\\n$", stdOut)
}
