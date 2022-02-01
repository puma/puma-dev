package dev

import (
	"strings"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestAppPool_findAppByDomainName(t *testing.T) {
	appLinkDir := homedir.MustExpand("~/.puma-dev-test_TestAppPool")

	testAppsToLink := map[string]string{
		"hipuma":     "rack-hi-puma",
		"yopuma":     "rack-hi-puma",
		"foo.hipuma": "static-hi-puma",
	}

	defer LinkTestApps(t, appLinkDir, testAppsToLink)()

	var testEvents Events
	testPool := &AppPool{
		Dir:    appLinkDir,
		Events: &testEvents,
	}

	_, err := testPool.FindAppByDomainName("doesnotexist")
	assert.Error(t, err)

	app_hipuma, _ := testPool.FindAppByDomainName("hipuma")
	assert.True(t, strings.HasPrefix(app_hipuma.Name, "rack-hi-puma-"))

	app_yopuma, _ := testPool.FindAppByDomainName("yopuma")
	assert.True(t, strings.HasPrefix(app_yopuma.Name, "rack-hi-puma-"))

	// foo.hipuma should match exactly and return static
	app_foo_hipuma, _ := testPool.FindAppByDomainName("foo.hipuma")
	assert.True(t, strings.HasPrefix(app_foo_hipuma.Name, "static-hi-puma-"))

	// foo.yopuma should strip foo and return rack
	app_foo_yopuma, _ := testPool.FindAppByDomainName("foo.yopuma")
	assert.True(t, strings.HasPrefix(app_foo_yopuma.Name, "rack-hi-puma-"))
}
