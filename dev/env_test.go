package dev

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

func TestLoadEnv_fileEnvPriority(t *testing.T) {
	base := make(map[string]string)
	appDir, _ := os.MkdirTemp("/tmp", "TestLoadEnv_fileEnvPriority")

	ioutil.WriteFile(filepath.Join(appDir, ".pumaenv"), []byte("FILE=.pumaenv"), fs.FileMode(0600))
	ioutil.WriteFile(filepath.Join(appDir, ".powenv"), []byte("FILE=.powenv"), fs.FileMode(0600))
	ioutil.WriteFile(filepath.Join(appDir, ".powrc"), []byte("FILE=.powrc"), fs.FileMode(0600))
	ioutil.WriteFile(filepath.Join(appDir, ".env"), []byte("FILE=.env"), fs.FileMode(0600))

	res, _ := LoadEnv(base, appDir)

	assert.Equal(t, res["FILE"], ".pumaenv")
}

func TestLoadEnv_homedirFileEnv(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping LoadEnv homedir tests outside CI")
	}

	base := make(map[string]string)
	appDir, _ := os.MkdirTemp("/tmp", "TestLoadEnv_homedirFileEnv")

	powconfig := homedir.MustExpand("~/.powconfig")
	pumaenv := homedir.MustExpand("~/.pumaenv")

	ioutil.WriteFile(powconfig, []byte("FILE=~/.powconfig"), fs.FileMode(0600))
	ioutil.WriteFile(pumaenv, []byte("FILE=~/.pumaenv"), fs.FileMode(0600))

	defer func() {
		os.Remove(powconfig)
		os.Remove(pumaenv)
	}()

	res, _ := LoadEnv(base, appDir)

	assert.Equal(t, res["FILE"], "~/.pumaenv")

}

func TestLoadEnv(t *testing.T) {
	t.Setenv("FILE", "environment")

	base := GetMapEnviron()
	appDir, _ := os.MkdirTemp("/tmp", "TestLoadEnv")

	res, _ := LoadEnv(base, appDir)

	assert.Equal(t, res["FILE"], "environment")
}
