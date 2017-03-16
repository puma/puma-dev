// +build darwin

package homedir

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func dir() (string, error) {
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
