package main

import (
	"testing"

	"github.com/puma/puma-dev/homedir"
)

func TestMainPumaDev_Linux(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.gotest-linux-puma-dev")

	defer linkAllTestApps(t, appLinkDir)()

	configureAndBootPumaDevServer(t, map[string]string{
		"dir":                   appLinkDir,
		"http-port":             "65080",
		"https-port":            "65443",
		"no-serve-public-paths": "/packs:/config",
	})

	runPlatformAgnosticTestScenarios(t)
}
