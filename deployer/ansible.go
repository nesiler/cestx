package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
	if !checkServiceExists(targetHost.Name) {
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

func checkServiceExists(host string) bool {
	playbook := config.AnsiblePath + "/check.yaml"
	err := runAnsiblePlaybook(playbook, host, map[string]string{"service": host})
	if err != nil {
		common.Err("Error running playbook 'check.yaml': %v", err)
		return false
	}

	// Read output from /tmp/check_result.txt
	file, err := os.Open("/tmp/check_result.txt")
	if err != nil {
		common.Err("Error opening check result file: %v", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file) // File contains just "True" or "False"
	scanner.Scan()
	result := scanner.Text() == "True"

	common.Warn("%s service exists: %v ", host, result)

	return result
}
