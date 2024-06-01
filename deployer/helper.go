package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	common.FailError(err, "error reading config file: %v")

	var config Config
	err = json.Unmarshal(data, &config)
	common.FailError(err, "error parsing config file: %v")

	return &config, nil
}

func exportSSHKeyToHost(keyName, identifierType, identifier string) error {
	keyPath := fmt.Sprintf("%s/.ssh/%s.pub", os.Getenv("HOME"), keyName)
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read the key file: %v", err)
	}

	proxmoxHost := os.Getenv("PROXMOX_HOST")
	if proxmoxHost == "" {
		return fmt.Errorf("PROXMOX_HOST is not set")
	}

	var url string
	var reqBody string

	switch identifierType {
	case "ip":
		url = fmt.Sprintf("http://%s:5252/ssh", proxmoxHost)
		reqBody = fmt.Sprintf(`{"ip": "%s", "key": "%s"}`, identifier, key)
	case "hostname":
		url = fmt.Sprintf("http://%s:5252/ssh", proxmoxHost)
		reqBody = fmt.Sprintf(`{"hostname": "%s", "key": "%s"}`, identifier, key)
	default:
		return fmt.Errorf("invalid identifier type: %s", identifierType)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return fmt.Errorf("failed to send key to proxmox container: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func setupSSHKeyForHost(keyName, hostName, ip string) error {
	if err := ensureSSHKeyExists(keyName); err != nil {
		return common.Err("failed to create SSH key: %v", err)
	}

	common.Info("Exporting SSH key to host %s\n", hostName)
	err := exportSSHKeyToHost(keyName, "ip", ip)
	if err != nil {
		common.Warn("Failed to export SSH key using IP for host %s: %v, trying hostname\n", hostName, err)
		err = exportSSHKeyToHost(keyName, "hostname", hostName)
		if err != nil {
			return common.Err("failed to export SSH key to host %s using both IP and hostname: %v", hostName, err)
		}
	} else {
		common.Ok("Successfully exported SSH key to host %s using IP: %s\n", hostName, ip)
	}

	return nil
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
