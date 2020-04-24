package dev

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestInstallIntoSystem(t *testing.T) {

	appLinkDir, _ := ioutil.TempDir("", ".puma-dev")
	libDir, _ := ioutil.TempDir("", "Library")
	logFilePath := filepath.Join(libDir, "Logs", "puma-dev.log")
	launchAgentDir := filepath.Join(libDir, "LaunchAgents")
	expectedPlistPath := filepath.Join(launchAgentDir, "io.puma.dev.plist")
	assert.NoDirExists(t, launchAgentDir)

	defer func() {
		exec.Command("launchctl", "unload", expectedPlistPath).Run()
		os.RemoveAll(appLinkDir)
		os.RemoveAll(logFilePath)
		os.RemoveAll(libDir)
	}()

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

	assert.DirExists(t, launchAgentDir)
	assert.FileExists(t, expectedPlistPath)
	AssertDirUmask(t, "0755", launchAgentDir)
}

func TestInstallIntoSystem_superuser(t *testing.T) {
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

func AssertDirUmask(t *testing.T, expectedUmask, path string) {
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
