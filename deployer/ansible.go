package main

import (
	"log"
	"os/exec"

	"github.com/nesiler/cestx/common"
)

func (c *Config) GetBuildCommand(serviceName string) string {
	return c.ServiceBuildCommands[serviceName]
}

// runAnsiblePlaybook runs an Ansible playbook
func runAnsiblePlaybook(playbookPath, host string) error {
	cmd := exec.Command("ansible-playbook", "-l", host, playbookPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

// Deploy deploys a service using Ansible
func Deploy(config *Config, serviceName string) error {
	inventoryPath := config.AnsiblePath + "/inventory.yaml"
	hosts, err := readInventory(inventoryPath)
	common.FailError(err, "Error reading inventory: %v")
	host := ""

	// Find host
	for name := range hosts {
		if name == serviceName {
			host = name
		}
	}

	// Check if repository is cloned
	err = runAnsiblePlaybook(config.AnsiblePath+"/check.yml", host)
	if err != nil {
		common.Err("Error checking repository and service: %v", err)
		return err
	}

	// Update or setup
	if checkServiceExists(host) {
		err = runAnsiblePlaybook(config.AnsiblePath+"/update.yml", host)
	} else {
		err = runAnsiblePlaybook(config.AnsiblePath+"/setup.yml", host)
	}
	if err != nil {
		common.Err("Error running playbook: %v", err)
		return err
	}

	return nil
}

// Check if service exists
func checkServiceExists(host string) bool {
	cmd := exec.Command("ansible", host, "-m", "systemd", "-a", "name={{ service }}", "-i", "ansible/inventory.yaml")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
