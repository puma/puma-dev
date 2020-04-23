package dev

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestInstallIntoSystem(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.gotest-dev-puma-dev")
	tmpDir := homedir.MustExpand("~/.gotest-dev-puma-dev-homedir")

	defer MakeDirectoryOrFail(t, appLinkDir)()
	defer MakeDirectoryOrFail(t, tmpDir)()

	launchAgentDir := filepath.Join(tmpDir, "Library", "LaunchAgents")
	assert.False(t, DirExists(launchAgentDir))

	err := InstallIntoSystem(&InstallIntoSystemArgs{
		ListenPort:         10080,
		TlsPort:            10443,
		Domains:            "test:localhost",
		Timeout:            "10s",
		ApplinkDirPath:     appLinkDir,
		LaunchAgentDirPath: launchAgentDir,
		LogfilePath:        "~/.gotest-dev-puma-dev-homedir/Library/Logs/puma-dev.log",
	})

	assert.NoError(t, err)

	info, err := os.Stat(launchAgentDir)
	if assert.NoError(t, err) {
		assert.Equal(t, 0755, info.Mode())
		assert.True(t, info.IsDir())
	}
}
