package dev

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

var AppEnvPaths = []string{".pumaenv", ".powenv", ".powrc", ".env"}
var GlobalEnvPaths = []string{"~/.powconfig", "~/.pumaenv"}

func Load(baseEnv map[string]string, dir string) (envMap map[string]string, err error) {
	// build a list of all the supported env files to load
	var envPaths []string

	for _, path := range AppEnvPaths {
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("Sourcing %s", fullPath)
			envPaths = append(envPaths, fullPath)
		}
	}

	for _, fullPath := range GlobalEnvPaths {
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("Sourcing %s", fullPath)
			envPaths = append(envPaths, fullPath)
		}
	}

	envMap = baseEnv

	// load the env into a map
	appEnv, err := godotenv.Read(envPaths...)
	if err != nil {
		return envMap, err
	}

	// merge contents from appEnv into envMap
	// this way dotenv values will supersede os env values
	for k, v := range appEnv {
		envMap[k] = v
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
