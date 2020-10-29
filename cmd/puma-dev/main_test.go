package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dev "github.com/puma/puma-dev/dev"
	. "github.com/puma/puma-dev/dev/devtest"

	"github.com/avast/retry-go"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	EnsurePumaDevDirectory()
	testResult := m.Run()
	os.Exit(testResult)
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

func configureAndBootPumaDevServer(t *testing.T, mainFlags map[string]string) error {
	StubCommandLineArgs()
	for flagName, flagValue := range mainFlags {
		SetFlagOrFail(t, flagName, flagValue)
	}

	address := fmt.Sprintf("localhost:%d", *fHTTPPort)
	timeout := time.Duration(2 * time.Second)

	if _, err := net.DialTimeout("tcp", address, timeout); err == nil {
		return fmt.Errorf("server is already running")
	}

	generateLivePumaDevCertIfNotExist(t)

	go func() {
		main()
	}()

	return retry.Do(
		func() error {
			_, err := net.DialTimeout("tcp", address, timeout)
			return err
		},
	)
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

func linkAllTestApps(t *testing.T, workingDirPath string) func() {
	MakeDirectoryOrFail(t, workingDirPath)

	testAppsToLink := map[string]string{
		"hipuma":      "rack-hi-puma",
		"static-site": "static-hi-puma",
	}

	for appLinkName, etcAppDir := range testAppsToLink {
		appPath := filepath.Join(ProjectRoot, "etc", etcAppDir)
		linkPath := filepath.Join(homedir.MustExpand(workingDirPath), appLinkName)

		if err := os.Symlink(appPath, linkPath); err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	return func() {
		RemoveDirectoryOrFail(t, workingDirPath)
	}
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

func runPlatformAgnosticTestScenarios(t *testing.T) {
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

	t.Run("static-site ignore packs", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/packs/site.js", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "rack wuz here", getURLWithHost(t, reqURL, statusHost))
	})

	t.Run("static-site ignore config", func(t *testing.T) {
		reqURL := fmt.Sprintf("http://localhost:%d/config.json", *fHTTPPort)
		statusHost := "static-site"

		assert.Equal(t, "rack wuz here", getURLWithHost(t, reqURL, statusHost))
	})
}
