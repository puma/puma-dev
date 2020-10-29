package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/puma/puma-dev/dev"
	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"

	"github.com/stretchr/testify/assert"
)

func TestMainPumaDev_Darwin(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.gotest-macos-puma-dev")

	defer linkAllTestApps(t, appLinkDir)()

	serveErr := configureAndBootPumaDevServer(t, map[string]string{
		"d":                     "test:puma",
		"dir":                   appLinkDir,
		"dns-port":              "65053",
		"http-port":             "65080",
		"https-port":            "65443",
		"no-serve-public-paths": "/packs:/config.json",
	})

	assert.NoError(t, serveErr)

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

		ips, err = r.LookupIPAddr(ctx, "foo.puma")
		assert.NoError(t, err)
		assert.Equal(t, net.ParseIP("127.0.0.1").To4(), ips[0].IP.To4())

		_, err = r.LookupIPAddr(ctx, "foo.tlddoesnotexist")
		assert.Error(t, err)
	})
}

func TestCertificateInstallAndTrustHTTPS_DarwinInteractive(t *testing.T) {
	if m := flag.Lookup("test.run").Value.String(); m == "" || !regexp.MustCompile(m).MatchString(t.Name()) {
		t.Skip("interactive test must be specified with -test.run=DarwinInteractive")
	}

	appLinkDir, _ := ioutil.TempDir("", ".puma-dev")

	liveSupportPath := homedir.MustExpand(dev.SupportDir)
	liveCertPath := filepath.Join(liveSupportPath, "cert.pem")
	liveKeyPath := filepath.Join(liveSupportPath, "key.pem")

	os.Remove(liveCertPath)
	os.Remove(liveKeyPath)

	assert.NoFileExists(t, liveCertPath)
	assert.NoFileExists(t, liveKeyPath)

	certInstallStdOut := WithStdoutCaptured(func() {
		err := dev.SetupOurCert()
		assert.Nil(t, err)

		assert.FileExists(t, liveCertPath)
		assert.FileExists(t, liveKeyPath)
	})

	defer func() {
		err := dev.DeleteAllPumaDevCAFromDefaultKeychain()
		assert.NoError(t, err)

		os.Remove(appLinkDir)
		os.Remove(liveCertPath)
		os.Remove(liveKeyPath)
	}()

	assert.Regexp(t, "^\\* Adding certification to login keychain as trusted\\n", certInstallStdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n", certInstallStdOut)
	assert.Regexp(t, "\\* Certificates setup, ready for https operations!\\n$", certInstallStdOut)

	defer linkAllTestApps(t, appLinkDir)()
	serveErr := configureAndBootPumaDevServer(t, map[string]string{
		"d":          "test:puma",
		"dir":        appLinkDir,
		"dns-port":   "55053",
		"http-port":  "55080",
		"https-port": "55443",
	})

	assert.NoError(t, serveErr)

	tlsRequestURL := fmt.Sprintf("https://localhost:%d/", *fTLSPort)
	tlsRequestHost := "hipuma"
	assert.Equal(t, "Hi Puma!", getURLWithHost(t, tlsRequestURL, tlsRequestHost))
}
