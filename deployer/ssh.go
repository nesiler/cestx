package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/nesiler/cestx/common"
)

func exportSSHKeyToHost(keyName, identifierType, identifier string) error {
	keyPath := fmt.Sprintf("%s/.ssh/%s.pub", os.Getenv("HOME"), keyName)
	key, err := os.ReadFile(keyPath)
	common.FailError(err, "failed to read SSH key: %v")

	proxmoxHost := os.Getenv("PROXMOX_HOST")
	if proxmoxHost == "" {
		return common.Err("PROXMOX_HOST environment variable is not set")
	}

	var url string
	var reqBody map[string]string

	switch identifierType {
	case "ip":
		url = fmt.Sprintf("http://%s:5252/ssh", proxmoxHost)
		reqBody = map[string]string{"ip": identifier, "key": string(key)}
	case "hostname":
		url = fmt.Sprintf("http://%s:5252/ssh", proxmoxHost)
		reqBody = map[string]string{"hostname": identifier, "key": string(key)}
	default:
		return common.Err("invalid identifier type: %s", identifierType)
	}

	jsonReqBody, err := json.Marshal(reqBody)
	common.FailError(err, "failed to marshal request body: %v")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	if err != nil {
		return common.Err("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return common.Err("failed to send key to proxmox container: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return common.Err("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func setupSSHKeyForHost(keyName, hostName, ip string) error {
	if err := ensureSSHKeyExists(keyName); err != nil {
		return common.Err("failed to create SSH key: %v", err)
	}

	common.Info("Exporting SSH key to host %s", hostName)
	err := exportSSHKeyToHost(keyName, "ip", ip)
	if err != nil {
		common.Warn("Failed to export SSH key using IP for host %s: %v, trying hostname", hostName, err)
		err = exportSSHKeyToHost(keyName, "hostname", hostName)
		if err != nil {
			return common.Err("failed to export SSH key to host %s using both IP and hostname: %v", hostName, err)
		}
	} else {
		common.Ok("Successfully exported SSH key to host %s using IP: %s", hostName, ip)
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

func checkSSHKeyExported(host string) bool {
	common.Info("Checking SSH key for host %s\n", host)
	cmd := exec.Command("ansible", host, "-m", "ping", "-i", "ansible/inventory.yaml", "--private-key", os.Getenv("HOME")+"/.ssh/master", "-u", "root")
	return cmd.Run() == nil
}
