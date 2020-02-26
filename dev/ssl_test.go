package dev

import (
	"crypto/tls"
	"path/filepath"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

func TestGeneratePumaDevCertificateAuthority(t *testing.T) {
	tmpPath := "tmp"

	defer MakeDirectoryOrFail(t, tmpPath)()

	testKeyPath := filepath.Join(tmpPath, "testkey.pem")
	testCertPath := filepath.Join(tmpPath, "testcert.pem")

	if err := GeneratePumaDevCertificateAuthority(testCertPath, testKeyPath); err != nil {
		assert.Fail(t, err.Error())
	}

	_, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)

	assert.NoError(t, err)
}
