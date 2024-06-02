package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/nesiler/cestx/common"
)

func (c *Config) GetBuildCommand(serviceName string) string {
	return c.ServiceBuildCommands[serviceName]
}

// runAnsiblePlaybook runs an Ansible playbook with optional extra variables
func runAnsiblePlaybook(playbookPath, host string, extraVars map[string]string) error {
	args := []string{"-l", host, playbookPath}
	for key, value := range extraVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

// Deploy deploys a service using Ansible
// Deploy deploys a service using Ansible
func Deploy(config *Config, serviceName string) error {
	inventoryPath := config.AnsiblePath + "/inventory.yaml"
	hosts, err := readInventory(inventoryPath)
	common.FailError(err, "Error reading inventory: %v")

	// Find host
	host, exists := hosts[serviceName]
	if !exists {
		return fmt.Errorf("host not found for service: %s", serviceName)
	}

	// Pass the service name as an extra variable to the playbook
	extraVars := map[string]string{"service": serviceName}

	// Check if repository and service exist
	err = runAnsiblePlaybook(config.AnsiblePath+"/check.yml", host, extraVars)
	if err != nil {
		common.Err("Error checking repository and service: %v", err)
		return err
	}

	// Update or setup
	playbook := config.AnsiblePath + "/update.yml"
	if !checkServiceExists(host, extraVars) {
		playbook = config.AnsiblePath + "/setup.yml"
	}

	err = runAnsiblePlaybook(playbook, host, extraVars)
	if err != nil {
		common.Err("Error running playbook: %v", err)
		return err
	}

	return nil
}

func checkServiceExists(host string, extraVars map[string]string) bool {
	args := []string{host, "-m", "systemd", "-a", "name={{ service }}", "-i", "ansible/inventory.yaml"}
	for key, value := range extraVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	cmd := exec.Command("ansible", args...)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
