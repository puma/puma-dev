package main

import (
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
)

func TestMainPumaDev_Linux(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.gotest-linux-puma-dev")

	defer linkAppsForTesting(t, appLinkDir)()

	SetFlagOrFail(t, "http-port", "65080")
	SetFlagOrFail(t, "https-port", "65443")

	bootConfiguredLivePumaServer(t, appLinkDir)

	runPlatformAgnosticTestScenarios(t)
}
