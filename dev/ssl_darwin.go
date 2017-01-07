package dev

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/puma/puma-dev/homedir"
)

const supportDir = "~/Library/Application Support/io.puma.dev"

func LoginKeyChain() (string, error) {

	new_keychain_path := homedir.MustExpand("~/Library/Keychains/login.keychain-db")
	old_keychain_path := homedir.MustExpand("~/Library/Keychains/login.keychain")

	if _, err := os.Stat(new_keychain_path); err == nil {
		return new_keychain_path, nil
	}

	if _, err := os.Stat(old_keychain_path); err == nil {
		return old_keychain_path, nil
	}

	return "", errors.New("Could not find login keychain")
}

func TrustCert(cert string) error {
	fmt.Printf("* Adding certification to login keychain as trusted\n")
	fmt.Printf("! There is probably a dialog open that requires you to authenticate\n")

	login, keychainError := LoginKeyChain()

	if keychainError != nil {
		return keychainError
	}

	err := exec.Command("sh", "-c",
		fmt.Sprintf(`security add-trusted-cert -d -r trustRoot -k '%s' '%s'`,
			login, cert)).Run()

	if err != nil {
		return err
	}

	fmt.Printf("* Certificates setup, ready for https operations!\n")

	return nil
}
