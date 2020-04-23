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
	appLinkDir := homedir.MustExpand("~/.gotest-macos-puma-dev-install-into-system")
	baseTmpDir := homedir.MustExpand("~/.gotest-puma-dev")

	defer MakeDirectoryOrFail(t, appLinkDir)()
	defer MakeDirectoryOrFail(t, baseTmpDir)()

	launchAgentDir := filepath.Join(baseTmpDir, "Library", "LaunchAgents")
	assert.False(t, DirExists(launchAgentDir))

	err := InstallIntoSystem(&InstallIntoSystemArgs{
		ApplinkDirPath:     appLinkDir,
		Domains:            "test:localhost",
		LaunchAgentDirPath: "~/.gotest/Library/LaunchAgents",
		ListenPort:         10080,
		LogfilePath:        "~/.gotest/Library/Logs",
		Timeout:            "10s",
		TlsPort:            10443,
	})

	assert.NoError(t, err)

	info, err := os.Stat(launchAgentDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}
