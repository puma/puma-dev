package dev

import "os/exec"

func Stop() error {
	err := exec.Command("pkill", "-USR1", "puma-dev").Run()
	if err != nil {
		return err
	}

	return nil
}
