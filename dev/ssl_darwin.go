package dev

import (
	"fmt"
	"os/exec"
)

const supportDir = "~/Library/Application Support/io.puma.dev"

func TrustCert(cert string) error {
	fmt.Printf("* Adding certification to login keychain as trusted\n")
	fmt.Printf("! There is probably a dialog open that you must type your password into\n")

	login := mustExpand("~/Library/Keychains/login.keychain")

	err := exec.Command("sh", "-c",
		fmt.Sprintf(`security add-trusted-cert -d -r trustRoot -k '%s' '%s'`,
			login, cert)).Run()

	if err != nil {
		return err
	}

	fmt.Printf("* Certificates setup, ready for https operations!\n")

	return nil
}
