package main

import (
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
	StubCommandLineArgs()

	SetFlagOrFail(t, "dir", testAppLinkDirPath)

	generateLivePumaDevCertIfNotExist(t)

	go func() {
		main()
	}()

	err := retry.Do(
		func() error {
			_, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", *fHTTPPort), time.Duration(10*time.Second))
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

func getUrlWithHost(t *testing.T, url string, host string) string {
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
			body := getUrlWithHost(t, fmt.Sprintf("http://localhost:%d/events", *fHTTPPort), "puma-dev")
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

func TestMainPumaDev(t *testing.T) {
	defer launchPumaDevBackgroundServerWithDefaults(t)()

	rackAppPath := filepath.Join(ProjectRoot, "etc", "rack-hi-puma")
	hipumaLinkPath := filepath.Join(homedir.MustExpand(testAppLinkDirPath), "hipuma")

	if err := os.Symlink(rackAppPath, hipumaLinkPath); err != nil {
		assert.FailNow(t, err.Error())
	}

	t.Run("status", func(t *testing.T) {
		statusUrl := fmt.Sprintf("http://localhost:%d/status", *fHTTPPort)
		statusHost := "puma-dev"

		assert.Equal(t, "{}", getUrlWithHost(t, statusUrl, statusHost))
	})

	t.Run("hipuma", func(t *testing.T) {
		appUrl := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		appHost := "hipuma"

		assert.Equal(t, "Hi Puma!", getUrlWithHost(t, appUrl, appHost))
	})

	t.Run("restart.txt", func(t *testing.T) {
		appUrl := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		appHost := "hipuma"
		appRestartTxt := filepath.Join(ProjectRoot, "etc", "rack-hi-puma", "tmp", "restart.txt")

		assert.Equal(t, "Hi Puma!", getUrlWithHost(t, appUrl, appHost))

		touchRestartTxt := exec.Command("sh", "-c", fmt.Sprintf("touch %s", appRestartTxt))

		if err := touchRestartTxt.Run(); err != nil {
			panic(err)
		}

		assert.NoError(t, pollForEvent(t, "rack-hi-puma", "killing_app", "restart.txt touched"))
		assert.NoError(t, pollForEvent(t, "rack-hi-puma", "shutdown", ""))
	})

	t.Run("unknown app", func(t *testing.T) {
		statusUrl := fmt.Sprintf("http://localhost:%d/", *fHTTPPort)
		statusHost := "doesnotexist"

		assert.Equal(t, "unknown app", getUrlWithHost(t, statusUrl, statusHost))
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
