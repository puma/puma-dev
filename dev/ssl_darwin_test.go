package dev

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func deleteAllPumaDevCAFromDefaultKeychain() {
	forEachPumaDevKeychainCertSha := func(subCommand string) string {
		return fmt.Sprintf(`for sha in $(security find-certificate -a -c "Puma-dev CA" -Z | awk '/SHA-1/ {print $3}'); do %s; done`, subCommand)
	}

	if err := exec.Command("sh", "-c", forEachPumaDevKeychainCertSha("security delete-certificate -Z $sha")).Run(); err != nil {
		panic(err)
	}

	log.Println("! NOTICE - REMOVED ALL CERTS LIKE \"Puma-dev CA\" FROM THE DEFAULT macOS KEYCHAIN")
}

func TestSetupOurCert_ensureNotWorldReadable(t *testing.T) {
	t.Skip("not implemented yet - https://github.com/puma/puma-dev/issues/215")
}

func TestSetupOurCert_InteractiveCertificateInstall(t *testing.T) {
	if flag.Lookup("test.run").Value.String() != t.Name() {
		t.Skipf("interactive test must be specified with -test.run=%s", t.Name())
	}

	liveSupportPath := homedir.MustExpand(SupportDir)
	liveCertPath := filepath.Join(liveSupportPath, "cert.pem")
	liveKeyPath := filepath.Join(liveSupportPath, "key.pem")

	os.Remove(liveCertPath)
	os.Remove(liveKeyPath)

	assert.False(t, FileExists(liveCertPath))
	assert.False(t, FileExists(liveKeyPath))

	certInstallStdOut := WithStdoutCaptured(func() {
		err := SetupOurCert()
		assert.Nil(t, err)

		assert.True(t, FileExists(liveCertPath))
		assert.True(t, FileExists(liveKeyPath))
	})

	assert.Regexp(t, "^\\* Adding certification to login keychain as trusted\\n", certInstallStdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n", certInstallStdOut)
	assert.Regexp(t, "\\* Certificates setup, ready for https operations!\\n$", certInstallStdOut)

	defer func() {
		deleteAllPumaDevCAFromDefaultKeychain()
		os.Remove(liveCertPath)
		os.Remove(liveKeyPath)
	}()
}

func TestTrustCert_Darwin_noCertProvided(t *testing.T) {
	stdOut := WithStdoutCaptured(func() {
		err := TrustCert("/does/not/exist")
		assert.NotNil(t, err)
		assert.Regexp(t, "Error reading file /does/not/exist\\n$", err)
	})

	assert.Regexp(t, "^* Adding certification to login keychain as trusted", stdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n$", stdOut)
}
