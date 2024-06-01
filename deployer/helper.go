package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
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
	common.FailError(err, "error reading config file: %v")

	var config Config
	err = json.Unmarshal(data, &config)
	common.FailError(err, "error parsing config file: %v")

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

func setupSSHKeysForHosts(keyName string, hosts []string) error {
	if err := ensureSSHKeyExists(keyName); err != nil {
		return common.Err("failed to create SSH key: %v", err)
	}

	for _, host := range hosts {
		if err := exportSSHKeyToHost(keyName, host); err != nil {
			return common.Err("failed to export SSH key to host %s: %v", host, err)
		}
	}

	return nil
}

func exportSSHKeyToHost(keyName, identifier string) error {
	keyPath := fmt.Sprintf("%s/.ssh/%s.pub", os.Getenv("HOME"), keyName)
	key, err := os.ReadFile(keyPath)
	common.FailError(err, "failed to read SSH key: %v", keyPath)

	godotenv.Load("../.env")
	proxmoxHost := os.Getenv("PROXMOX_HOST")
	if proxmoxHost == "" {
		return common.Err("PROXMOX_HOST environment variable not set")
	}

	// Attempt to export using IP address
	url := fmt.Sprintf("http://%s:5252/ssh", proxmoxHost)
	reqBody := fmt.Sprintf(`{"ip": "%s", "key": "%s"}`, identifier, key)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}

	// If IP address export fails, attempt to export using hostname
	reqBody = fmt.Sprintf(`{"hostname": "%s", "key": "%s"}`, identifier, key)
	resp, err = http.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}

	common.FailError(err, "failed to export SSH key to host %s: %v", identifier, err)

	return common.Err("unexpected status code: %d", resp.StatusCode)
}
