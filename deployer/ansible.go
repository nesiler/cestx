package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/nesiler/cestx/common"
)

func GetBuildCommand(serviceName string) string {
	return config.ServiceBuildCommands[serviceName]
}

// runAnsiblePlaybook runs an Ansible playbook with optional extra variables
func runAnsiblePlaybook(playbookPath, host string, extraVars map[string]string) error {
	args := []string{"-l", host, playbookPath}
	for key, value := range extraVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// set inventory file
	args = append(args, "-i", config.AnsiblePath+"/inventory.yaml")

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

// Deploy deploys a service using Ansible
func Deploy(serviceName string) error {
	inventoryPath := config.AnsiblePath + "/inventory.yaml"
	hosts, err := readInventory(inventoryPath)
	common.FailError(err, "Error reading inventory: %v")

	// Find host by service name (assuming service name matches host name in inventory)
	var targetHost *Host
	for _, h := range hosts {
		if h.Name == serviceName {
			targetHost = &h
			break
		}
	}

	if targetHost == nil {
		return common.Err("Error: Host not found for service %s", serviceName)
	}

	// Pass the service name as an extra variable to the playbook
	extraVars := map[string]string{"service": serviceName}
	playbook := config.AnsiblePath + "/update.yaml"

	// Check if repository and service exist
	repoExists, serviceExists := checkServiceExists(targetHost.AnsibleHost, extraVars) // Pass the host's IP address
	if !repoExists || !serviceExists {
		playbook = config.AnsiblePath + "/setup.yaml"
	}

	err = runAnsiblePlaybook(playbook, targetHost.AnsibleHost, extraVars)
	if err != nil {
		return common.Err("Error running playbook: %v", err)
	}

	return nil
}

func checkServiceExists(host string, extraVars map[string]string) (bool, bool) { // Now returns two booleans
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := []string{"-l", host, config.AnsiblePath + "/setup.yaml"} // Assuming the playbook is named check.yml
	for key, value := range extraVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		common.Err("Error running Ansible check playbook: %v", err)
		common.Err("Stderr: %s", stderr.String())
		return false, false
	}

	output := stdout.String()
	repoExists := strings.Contains(output, "Repository exists: True")
	serviceActive := strings.Contains(output, "Service is active: True")

	return repoExists, serviceActive
}
