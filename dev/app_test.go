package dev

import (
	"strings"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestAppPool_FindAppByDomainName_returnsAppWithExactName(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.puma-dev-test_app-test-exact-name-puma-dev")
	testAppsToLink := map[string]string{
		"hipuma": "rack-hi-puma",
	}
	defer LinkTestApps(t, appLinkDir, testAppsToLink)()

	var events Events

	var testPool AppPool
	testPool.Dir = appLinkDir
	testPool.Events = &events

	app, err := testPool.FindAppByDomainName("hipuma")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.True(t, strings.HasPrefix(app.Name, "rack-hi-puma-"))
}

func TestAppPool_FindAppByDomainName_returnsAppWithExactSubdomain(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.puma-dev-test_app-test-exact-subdomain-puma-dev")
	testAppsToLink := map[string]string{
		"hipuma":     "rack-hi-puma",
		"foo.hipuma": "static-hi-puma",
	}
	defer LinkTestApps(t, appLinkDir, testAppsToLink)()

	var events Events

	var testPool AppPool
	testPool.Dir = appLinkDir
	testPool.Events = &events

	app, err := testPool.FindAppByDomainName("foo.hipuma")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.True(t, strings.HasPrefix(app.Name, "static-hi-puma-"))
}

func TestAppPool_FindAppByDomainName_returnsAppWithBaseDomain(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.puma-dev-test_app-test-base-subdomain-puma-dev")
	testAppsToLink := map[string]string{
		"hipuma": "rack-hi-puma",
	}
	defer LinkTestApps(t, appLinkDir, testAppsToLink)()

	var events Events

	var testPool AppPool
	testPool.Dir = appLinkDir
	testPool.Events = &events

	app, err := testPool.FindAppByDomainName("foo.hipuma")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.True(t, strings.HasPrefix(app.Name, "rack-hi-puma-"))
}
