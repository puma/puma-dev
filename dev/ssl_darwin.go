package dev

import (
	"bytes"
	"fmt"
	"os/exec"
)

// SupportDir is the platform-specific path that contains puma-dev's generated certs
const SupportDir = "~/Library/Application Support/io.puma.dev"

// TrustCert adds the cert at the provided path to the macOS default login keychain
func TrustCert(cert string) error {
	fmt.Printf("* Adding certification to login keychain as trusted\n")
	fmt.Printf("! There is probably a dialog open that requires you to authenticate\n")

	login, keychainError := loginKeyChain()

	if keychainError != nil {
		return keychainError
	}

	addTrustedCertCommand := exec.Command("sh", "-c", fmt.Sprintf(`security add-trusted-cert -k '%s' '%s'`, login, cert))

	var stderr bytes.Buffer
	addTrustedCertCommand.Stderr = &stderr

	if err := addTrustedCertCommand.Run(); err != nil {
		return fmt.Errorf("add-trusted-cert had %s. %s", err.Error(), stderr.Bytes())
	}

	fmt.Printf("* Certificates setup, ready for https operations!\n")

	return nil
}

func DeleteAllPumaDevCAFromDefaultKeychain() error {
	deleteAllBashCommand := `
	for sha in $(security find-certificate -a -c "Puma-dev CA" -Z | awk '/SHA-1/ {print $3}'); do 
		security delete-certificate -t -Z $sha || security delete-certificate -Z $sha
	done
	`

	deleteAllPumaDevTrustsCmd := exec.Command("sh", "-c", deleteAllBashCommand)

	var stdout, stderr bytes.Buffer
	deleteAllPumaDevTrustsCmd.Stdout = &stdout
	deleteAllPumaDevTrustsCmd.Stderr = &stderr

	if err := deleteAllPumaDevTrustsCmd.Run(); err != nil {
		return err
	}

	return nil
}

func loginKeyChain() (string, error) {
	discoverLoginKeychainCmd := exec.Command("sh", "-c", `security login-keychain | xargs | tr -d '"' | tr -d '\n'`)

	var stdout, stderr bytes.Buffer
	discoverLoginKeychainCmd.Stdout = &stdout
	discoverLoginKeychainCmd.Stderr = &stderr

	if err := discoverLoginKeychainCmd.Run(); err != nil {
		return "", fmt.Errorf("could not find login keychain. security login-keychain had %s, %s", err.Error(), stderr.Bytes())
	}

	return string(stdout.Bytes()), nil
}
