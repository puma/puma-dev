package dev

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Setup(skipFirewall bool) error {
	err := os.MkdirAll(etcDir, 0755)
	if err != nil {
		return err
	}

	var ok bool

	sudo := os.Getenv("SUDO_USER")
	if sudo != "" {
		uid, err1 := strconv.Atoi(os.Getenv("SUDO_UID"))
		gid, err2 := strconv.Atoi(os.Getenv("SUDO_GID"))

		if err1 == nil && err2 == nil {
			fmt.Printf("* Configuring %s to be owned by %s\n", etcDir, sudo)

			err := os.Chown(etcDir, uid, gid)
			if err != nil {
				return err
			}

			err = os.Chmod(etcDir, 0755)
			if err != nil {
				return err
			}

			ok = true
		}
	}

	if !ok {
		fmt.Printf("* Configuring %s to be world writable\n")
		err := os.Chmod(etcDir, 0777)
		if err != nil {
			return err
		}
	}

	if !skipFirewall {
		fmt.Printf("* Configuring firewall...\n")

		cmd := exec.Command("pfctl", "-a", "com.apple/250.PumaDevFirewall", "-E", "-f", "-")
		rule := fmt.Sprintf(
			"rdr pass inet proto tcp from any to any port = %d -> 127.0.0.1 port %d\n",
			80, 9280)

		cmd.Stdin = strings.NewReader(rule)

		err = cmd.Run()
		if err != nil {
			return err
		}

		cur, err := exec.Command(
			"pfctl", "-a", "com.apple/250.PumaDevFirewall", "-s", "nat", "-q").Output()
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(cur)) != strings.TrimSpace(rule) {
			return fmt.Errorf("Unable to verify firewall installation was successful")
		}
	}

	return nil
}
