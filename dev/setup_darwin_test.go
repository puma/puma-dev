package dev

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestInstallIntoSystem_VerifyLaunchAgent(t *testing.T) {
	launchAgentDir, expectedPlistPath, cleanupFunc := installIntoTestContext(t)
	defer cleanupFunc()

	assert.FileExists(t, expectedPlistPath)
	assertDirUmask(t, "0755", launchAgentDir)

	assert.NoError(t, exec.Command("launchctl", "list", "io.puma.dev").Run())
}

func TestUninstall_DarwinInteractive(t *testing.T) {
	if m := flag.Lookup("test.run").Value.String(); m == "" || !regexp.MustCompile(m).MatchString(t.Name()) {
		t.Skip("interactive test must be specified with -test.run=DarwinInteractive")
	}

	launchAgentDir, _, cleanupFunc := installIntoTestContext(t)
	defer cleanupFunc()

	assert.NoError(t, exec.Command("launchctl", "list", "io.puma.dev").Run())

	Uninstall(launchAgentDir, []string{"test", "localhost"})

	assert.Error(t, exec.Command("launchctl", "list", "io.puma.dev").Run())
	assert.NoDirExists(t, SupportDir)
}

func TestInstallIntoSystem_FailsAsSuperuser(t *testing.T) {
	os.Setenv("SUDO_USER", os.Getenv("USER"))
	defer os.Unsetenv("SUDO_USER")

	err := InstallIntoSystem(&InstallIntoSystemArgs{
		ListenPort:         10080,
		TlsPort:            10443,
		Domains:            "test:localhost",
		Timeout:            "5s",
		ApplinkDirPath:     "/tmp/gotest-dummy-applinkdir",
		LaunchAgentDirPath: "/tmp/gotest-dummy-launchagent",
		LogfilePath:        "/tmp/gotest-dummy-logs/dummy.log",
	})

	assert.Error(t, err)
}

func installIntoTestContext(t *testing.T) (string, string, func()) {
	appLinkDir, _ := ioutil.TempDir("", ".puma-dev")
	libDir, _ := ioutil.TempDir("", "Library")
	logFilePath := filepath.Join(libDir, "Logs", "puma-dev.log")
	launchAgentDir := filepath.Join(libDir, "LaunchAgents")
	assert.NoDirExists(t, launchAgentDir)

	expectedPlistPath := filepath.Join(launchAgentDir, "io.puma.dev.plist")

	cleanup := func() {
		exec.Command("launchctl", "unload", expectedPlistPath).Run()
		os.RemoveAll(appLinkDir)
		os.RemoveAll(logFilePath)
		os.RemoveAll(libDir)
	}

	generateLivePumaDevCertIfNotExist(t)

	err := InstallIntoSystem(&InstallIntoSystemArgs{
		ListenPort:         10080,
		TlsPort:            10443,
		Domains:            "test:localhost",
		Timeout:            "5s",
		ApplinkDirPath:     appLinkDir,
		LaunchAgentDirPath: launchAgentDir,
		LogfilePath:        logFilePath,
	})
	assert.NoError(t, err)

	return launchAgentDir, expectedPlistPath, cleanup
}

func assertDirUmask(t *testing.T, expectedUmask, path string) {
	info, err := os.Stat(path)
	if !assert.NoError(t, err) {
		assert.True(t, info.IsDir())

		actualUmask := "0" + strconv.FormatInt(int64(info.Mode().Perm()), 8)
		assert.Equal(t, expectedUmask, actualUmask)
	}
}

func generateLivePumaDevCertIfNotExist(t *testing.T) {
	liveSupportPath := homedir.MustExpand(SupportDir)
	liveCertPath := filepath.Join(liveSupportPath, "cert.pem")
	liveKeyPath := filepath.Join(liveSupportPath, "key.pem")

	if !FileExists(liveCertPath) || !FileExists(liveKeyPath) {
		MakeDirectoryOrFail(t, liveSupportPath)

		if err := GeneratePumaDevCertificateAuthority(liveCertPath, liveKeyPath); err != nil {
			assert.FailNow(t, err.Error())
		}
	}
}
