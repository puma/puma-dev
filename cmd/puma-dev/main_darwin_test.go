package main

import (
	"context"
	"fmt"
	"net"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"

	"github.com/stretchr/testify/assert"
)

func TestMainPumaDev_Darwin(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.gotest-macos-puma-dev")

	defer linkAllTestApps(t, appLinkDir)()

	configureAndBootPumaDevServer(t, map[string]string{
		"dir":        appLinkDir,
		"dns-port":   "65053",
		"http-port":  "65080",
		"https-port": "65443",
	})

	runPlatformAgnosticTestScenarios(t)

	t.Run("resolve dns", func(t *testing.T) {
		PumaDevDNSDialer := func(ctx context.Context, network, address string) (net.Conn, error) {
			dnsAddress := fmt.Sprintf("127.0.0.1:%d", *fDNSPort)
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", dnsAddress)
		}

		r := net.Resolver{
			PreferGo: true,
			Dial:     PumaDevDNSDialer,
		}

		ctx := context.Background()
		ips, err := r.LookupIPAddr(ctx, "foo.test")

		assert.NoError(t, err)
		assert.Equal(t, net.ParseIP("127.0.0.1").To4(), ips[0].IP.To4())
	})
}
