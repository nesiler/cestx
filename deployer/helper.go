package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/nesiler/cestx/common"
)

type Config struct {
	GitHubToken   string `json:"github_token"`
	RepoOwner     string `json:"repo_owner"`
	RepoName      string `json:"repo_name"`
	RepoPath      string `json:"repo_path"`
	CheckInterval int    `json:"check_interval"`
	AnsiblePath   string `json:"ansible_path"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

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
		return common.Fatal("failed to create SSH key: %v", err)
	}

	for _, host := range hosts {
		if err := exportSSHKeyToHost(keyName, user, host); err != nil {
			return fmt.Errorf("failed to export SSH key to host %s: %w", host, err)
		}
	}

	return nil
}
