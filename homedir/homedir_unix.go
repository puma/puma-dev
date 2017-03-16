// +build !darwin,!windows

package homedir

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func dir() (string, error) {
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
