package dev

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/puma/puma-dev/homedir"
)

const supportDir = "~/Library/Application Support/io.puma.dev"

func loginKeyChain() (string, error) {

	newKeychainPath := homedir.MustExpand("~/Library/Keychains/login.keychain-db")
	oldKeychainPath := homedir.MustExpand("~/Library/Keychains/login.keychain")

	if _, err := os.Stat(newKeychainPath); err == nil {
		return newKeychainPath, nil
	}

	if _, err := os.Stat(oldKeychainPath); err == nil {
		return oldKeychainPath, nil
	}

	return "", errors.New("Could not find login keychain")
}

func trustCert(cert string) error {
	fmt.Printf("* Adding certification to login keychain as trusted\n")
	fmt.Printf("! There is probably a dialog open that requires you to authenticate\n")

	login, keychainError := loginKeyChain()

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
