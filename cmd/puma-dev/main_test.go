package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	dev "github.com/puma/puma-dev/dev"
	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/avast/retry-go"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

var testAppLinkDirPath = homedir.MustExpand("~/.gotest-puma-dev")

func TestMain(m *testing.M) {
	EnsurePumaDevDirectory()
	os.Exit(m.Run())
}

func TestMainPumaDev(t *testing.T) {
	defer launchPumaDevBackgroundServerWithDefaults(t)()

	testAppsToLink := map[string]string{
		"rack-hi-puma":   "hipuma",
		"static-hi-puma": "static-site",
	}

	for etcAppDir, appLinkName := range testAppsToLink {
		appPath := filepath.Join(ProjectRoot, "etc", etcAppDir)
		linkPath := filepath.Join(homedir.MustExpand(testAppLinkDirPath), appLinkName)

		if err := os.Symlink(appPath, linkPath); err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	t.Run("resolve dns", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.SkipNow()
		}

		PumaDevDNSDialer := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			// TODO: use fPort if available, break out this test into main_darwin_test.go
			return d.DialContext(ctx, "udp", "127.0.0.1:9253")
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

	t.Run("status", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/status", *fHTTPPort)
		statusHost := "puma-dev"

		assert.Equal(t, "{}", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("hipuma rack response", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		reqHost := "hipuma"

		assert.Equal(t, "Hi Puma!", getURLWithHost(t, reqURL, reqHost))
	})

	t.Run("hipuma tmp/restart.txt", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		reqHost := "hipuma"
		appRestartTxt := filepath.Join(ProjectRoot, "etc", "rack-hi-puma", "tmp", "restart.txt")

		assert.Equal(t, "Hi Puma!", getURLWithHost(t, reqURL, reqHost))

		touchRestartTxt := exec.Command("sh", "-c", fmt.Sprintf("touch %s", appRestartTxt))

		if err := touchRestartTxt.Run(); err != nil {
			panic(err)
		}

		assert.NoError(t, pollForEvent(t, "rack-hi-puma", "killing_app", "restart.txt touched"))
		assert.NoError(t, pollForEvent(t, "rack-hi-puma", "shutdown", ""))
	})

	t.Run("unknown app", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		statusHost := "doesnotexist"

		assert.Equal(t, "unknown app", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site public/greeting.txt", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/greeting.txt", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "Hi Puma!", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site public/index.html", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/index.html", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "<html><h1>public/index.html</h1></html>", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site public/subfolder/index.html", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/subfolder/index.html", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "<html><h1>public/subfolder/index.html</h1></html>", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site public/subfolder/../index.html", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/subfolder/../index.html", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "<html><h1>public/index.html</h1></html>", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site public/../index.html", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/../index.html", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "<html><h1>public/index.html</h1></html>", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site public/subfolder/../../index.html", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/subfolder/../../index.html", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "<html><h1>public/index.html</h1></html>", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site rack response", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/foo", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "rack wuz here", getURLWithHost(t, reqURL, statusHost))
	})
}

func TestMain_execWithExitStatus_versionFlag(t *testing.T) {
	StubCommandLineArgs("-V")
	assert.True(t, *fVersion)

	execStdOut := WithStdoutCaptured(func() {
		result := execWithExitStatus()
		assert.Equal(t, 0, result.exitStatusCode)
		assert.Equal(t, true, result.shouldExit)
	})

	assert.Regexp(t, "^Version: devel \\(.+\\)\\n$", execStdOut)
}

func TestMain_execWithExitStatus_noFlag(t *testing.T) {
	StubCommandLineArgs()
	assert.False(t, *fVersion)

	execStdOut := WithStdoutCaptured(func() {
		result := execWithExitStatus()
		assert.Equal(t, false, result.shouldExit)
	})

	assert.Equal(t, "", execStdOut)
}

func TestMain_execWithExitStatus_commandArgs(t *testing.T) {
	StubCommandLineArgs("nosoupforyou")

	execStdOut := WithStdoutCaptured(func() {
		result := execWithExitStatus()
		assert.Equal(t, 1, result.exitStatusCode)
		assert.Equal(t, true, result.shouldExit)
	})

	assert.Equal(t, "Error: Unknown command: nosoupforyou\n\n", execStdOut)
}

func TestMain_allCheck_versionFlag(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubCommandLineArgs("-V")
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_versionFlag")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	assert.Nil(t, err)
}

func TestMain_allCheck_badArg(t *testing.T) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		StubCommandLineArgs("-badarg")
		allCheck()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_allCheck_badArg")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()

	exit, ok := err.(*exec.ExitError)

	assert.True(t, ok)
	assert.False(t, exit.Success())
}

func generateLivePumaDevCertIfNotExist(t *testing.T) {
	liveSupportPath := homedir.MustExpand(dev.SupportDir)
	liveCertPath := filepath.Join(liveSupportPath, "cert.pem")
	liveKeyPath := filepath.Join(liveSupportPath, "key.pem")

	if !FileExists(liveCertPath) || !FileExists(liveKeyPath) {
		MakeDirectoryOrFail(t, liveSupportPath)

		if err := dev.GeneratePumaDevCertificateAuthority(liveCertPath, liveKeyPath); err != nil {
			assert.FailNow(t, err.Error())
		}
	}
}

func launchPumaDevBackgroundServerWithDefaults(t *testing.T) func() {
	address := fmt.Sprintf("localhost:%d", *fHTTPPort)
	timeout := time.Duration(10 * time.Second)

	if _, err := net.DialTimeout("tcp", address, timeout); err == nil {
		return func() {}
	}

	StubCommandLineArgs()
	SetFlagOrFail(t, "dir", testAppLinkDirPath)
	generateLivePumaDevCertIfNotExist(t)

	go func() {
		main()
	}()

	err := retry.Do(
		func() error {
			_, err := net.DialTimeout("tcp", address, timeout)
			return err
		},
		retry.OnRetry(func(n uint, err error) {
			log.Printf("#%d: %s\n", n, err)
		}),
	)

	assert.NoError(t, err)

	return func() {
		RemoveDirectoryOrFail(t, testAppLinkDirPath)
	}
}

func getURLWithHost(t *testing.T, url string, host string) string {
	req, _ := http.NewRequest("GET", url, nil)
	req.Host = host

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	return strings.TrimSpace(string(bodyBytes))
}

func pollForEvent(t *testing.T, app string, event string, reason string) error {
	return retry.Do(
		func() error {
			body := getURLWithHost(t, fmt.Sprintf("http://localhost:%d/events", *fHTTPPort), "puma-dev")
			eachEvent := strings.Split(body, "\n")

			for _, line := range eachEvent {
				var rawEvt interface{}
				if err := json.Unmarshal([]byte(line), &rawEvt); err != nil {
					assert.FailNow(t, err.Error())
				}
				evt := rawEvt.(map[string]interface{})
				LogDebugf("%+v", evt)

				eventFound := (app == "" || evt["app"] == app) &&
					(event == "" || evt["event"] == event) &&
					(reason == "" || evt["reason"] == reason)

				if eventFound {
					return nil
				}
			}

			return errors.New("not found")
		},
		retry.RetryIf(func(err error) bool {
			return err.Error() == "not found"
		}),
	)
}
