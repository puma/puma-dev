// https://github.com/mitchellh/go-homedir

package homedir

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// DisableCache will disable caching of the home directory. Caching is enabled
// by default.
var DisableCache bool

var homedirCache string
var cacheLock sync.RWMutex

var ErrNoHomeDir = errors.New("no home directory available")

// Dir returns the home directory for the executing user.
//
// This uses an OS-specific method for discovering the home directory.
// An error is returned if a home directory cannot be detected.
func Dir() (string, error) {
	if !DisableCache {
		cacheLock.RLock()
		cached := homedirCache
		cacheLock.RUnlock()
		if cached != "" {
			return cached, nil
		}
	}

	cacheLock.Lock()
	defer cacheLock.Unlock()

	var result string
	var err error
	switch runtime.GOOS {
	case "windows":
		result, err = dirWindows()
	case "darwin":
		result, err = dirDarwin()
	default:
		// Unix-like system, so just assume Unix
		result, err = dirUnix()
	}

	if err != nil {
		return "", err
	}

	homedirCache = result
	return result, nil
}

// Expand expands the path to include the home directory if the path
// is prefixed with `~`. If it isn't prefixed with `~`, the path is
// returned as-is.
func Expand(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", errors.New("cannot expand user-specific home dir")
	}

	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, path[1:]), nil
}

func MustExpand(path string) string {
	str, err := Expand(path)
	if err != nil {
		panic(err)
	}

	return str
}

func dirDarwin() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	var stdout bytes.Buffer

	// If that fails, try OS specific commands
	cmd := exec.Command("sh", "-c", `dscl -q . -read /Users/"$(whoami)" NFSHomeDirectory | sed 's/^[^ ]*: //'`)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		result := strings.TrimSpace(stdout.String())
		if result != "" {
			return result, nil
		}
	}

	// try the shell
	stdout.Reset()
	cmd = exec.Command("sh", "-c", "cd && pwd")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		result := strings.TrimSpace(stdout.String())
		if result != "" {
			return result, nil
		}
	}

	// try to figure out the user and check the default location
	stdout.Reset()
	cmd = exec.Command("whoami")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		user := strings.TrimSpace(stdout.String())

		path := "/Users/" + user

		stat, err := os.Stat(path)
		if err == nil && stat.IsDir() {
			return path, nil
		}
	}

	return "", ErrNoHomeDir
}

func dirUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	var stdout bytes.Buffer

	// If that fails, try OS specific commands
	cmd := exec.Command("getent", "passwd", strconv.Itoa(os.Getuid()))
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		if passwd := strings.TrimSpace(stdout.String()); passwd != "" {
			// username:password:uid:gid:gecos:home:shell
			passwdParts := strings.SplitN(passwd, ":", 7)
			if len(passwdParts) > 5 {
				return passwdParts[5], nil
			}
		}
	}

	// If all else fails, try the shell
	stdout.Reset()
	cmd = exec.Command("sh", "-c", "cd && pwd")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		result := strings.TrimSpace(stdout.String())
		if result == "" {
			return "", errors.New("blank output when reading home directory")
		}
	}

	// try to figure out the user and check the default location
	stdout.Reset()
	cmd = exec.Command("whoami")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		user := strings.TrimSpace(stdout.String())

		path := "/home/" + user

		stat, err := os.Stat(path)
		if err == nil && stat.IsDir() {
			return path, nil
		}
	}

	return "", ErrNoHomeDir
}

func dirWindows() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}
