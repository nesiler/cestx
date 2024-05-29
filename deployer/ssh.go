package main

import (
	"fmt"
	"os"
	"os/exec"
)

func ensureSSHKeyExists(keyName string) error {
	keyPath := fmt.Sprintf("%s/.ssh/%s", os.Getenv("HOME"), keyName)
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// Create the SSH key
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", keyPath, "-N", "")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

func exportSSHKeyToHost(keyName, user, host string) error {
	keyPath := fmt.Sprintf("%s/.ssh/%s.pub", os.Getenv("HOME"), keyName)
	authorizedKeysCmd := fmt.Sprintf("ssh %s@%s 'grep -q \"$(cat %s)\" ~/.ssh/authorized_keys || cat >> ~/.ssh/authorized_keys'", user, host, keyPath)

	// Check if the key is already in authorized_keys
	cmd := exec.Command("sh", "-c", authorizedKeysCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setupSSHKeysForHosts(keyName string, user string, hosts []string) error {
	if err := ensureSSHKeyExists(keyName); err != nil {
		return fmt.Errorf("failed to create SSH key: %w", err)
	}

	for _, host := range hosts {
		if err := exportSSHKeyToHost(keyName, user, host); err != nil {
			return fmt.Errorf("failed to export SSH key to host %s: %w", host, err)
		}
	}

	return nil
}
