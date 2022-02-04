package dev

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/puma/puma-dev/homedir"
)

// load paths in order of preference, first to set a variable wins
// `.pumaenv` has highest priority, `~/.powconfig` has lowest.
var AppEnvPaths = []string{".env", ".powrc", ".powenv", ".pumaenv"}
var HomeEnvPaths = []string{"~/.powconfig", "~/.pumaenv"}

func LoadEnv(baseEnv map[string]string, dir string) (envMap map[string]string, err error) {
	// build a list of all the supported env files that exist on the filesystem to load
	var envPaths []string

	for _, path := range AppEnvPaths {
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			envPaths = append(envPaths, fullPath)
		}
	}

	for _, path := range HomeEnvPaths {
		if fullPath, err := homedir.Expand(path); err == nil {
			if _, err := os.Stat(fullPath); err == nil {
				envPaths = append(envPaths, fullPath)
			}
		}
	}

	envMap = baseEnv

	// if we have any optional env paths to source, source them
	if len(envPaths) > 0 {
		// load the env into a map
		appEnv, err := godotenv.Read(envPaths...)
		if err != nil {
			// if any path fails to load, bail out.
			return nil, fmt.Errorf("LoadEnv failed. %s", err.Error())
		}

		// merge contents from appEnv into envMap
		// this way dotenv values will supersede os env values
		for k, v := range appEnv {
			envMap[k] = v
		}
	}

	return
}

func ToCmdEnv(envMap map[string]string) (env []string) {
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

func GetMapEnviron() (envMap map[string]string) {
	envMap = make(map[string]string)

	for _, keyval := range os.Environ() {
		pair := strings.Split(keyval, "=")
		envMap[pair[0]] = pair[1]
	}

	return
}
