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

	common.Out("Running Ansible playbook: ansible-playbook %s", strings.Join(args, " "))
	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

// Deploy deploys a service using Ansible
func Deploy(serviceName string) error {

	// Find host by service name (assuming service name matches host name in inventory)
	var targetHost *Host
	for _, h := range hosts {
		if h.Name == serviceName {
			targetHost = &h
			common.Ok("Found host for service %s: %s", serviceName, targetHost.AnsibleHost)
			break
		}
	}

	if targetHost == nil {
		return common.Err("Error: Host not found for service %s", serviceName)
	}

	playbook := config.AnsiblePath + "/update.yaml"

	// Check if repository and service exist
	repoExists, serviceExists := checkServiceExists(targetHost.Name) // Pass the host's IP address
	if !repoExists || !serviceExists {
		common.Warn("Service or Repo does not exist for host %s\n", targetHost.Name)
		common.Info("Starting setup process for: %s\n", targetHost.Name)
		playbook = config.AnsiblePath + "/setup.yaml"
	}

	err := runAnsiblePlaybook(playbook, targetHost.AnsibleHost, map[string]string{"service": serviceName})
	if err != nil {
		return common.Err("Error running playbook: %v", err)
	}

	return nil
}

func checkServiceExists(host string) (bool, bool) { // Now returns two booleans
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := []string{"-l", host, config.AnsiblePath + "/check.yaml"} // Assuming the playbook is named check.yml
	args = append(args, "-e", "service="+host)
	args = append(args, "-i", config.AnsiblePath+"/inventory.yaml")

	common.Out("Running Ansible check playbook: ansible-playbook %s", strings.Join(args, " "))
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
	// common.Warn("Ansible check output: %s", output)
	repoExists := strings.Contains(output, "Repository exists: True")
	serviceActive := strings.Contains(output, "Service is active: True")

	return repoExists, serviceActive
}
