package main

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
	"gopkg.in/yaml.v2"
)

// Host represents a single host in the inventory.
type Host struct {
	Name                     string `yaml:"name"`
	AnsibleHost              string `yaml:"ansible_host"`
	AnsibleSSHPrivateKeyFile string `yaml:"ansible_ssh_private_key_file"`
}

// Inventory represents the structure of the inventory YAML file.
type Inventory struct {
	Services struct {
		Hosts map[string]Host `yaml:"hosts"`
	} `yaml:"services"`
}

type Config struct {
	GitHubToken          string            `json:"github_token"`
	RepoOwner            string            `json:"repo_owner"`
	RepoName             string            `json:"repo_name"`
	RepoPath             string            `json:"repo_path"`
	CheckInterval        int               `json:"check_interval"`
	AnsiblePath          string            `json:"ansible_path"`
	ServiceBuildCommands map[string]string `json:"service_build_commands"`
}

var (
	config *Config
	hosts  []Host
)

func LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	common.FailError(err, "error reading config file: %v")

	err = json.Unmarshal(data, &config)
	common.FailError(err, "error parsing config file: %v")

	return nil
}

func readInventory(filePath string) error {
	data, err := os.ReadFile(filePath)
	common.FailError(err, "error reading inventory file: %v", filePath)

	// Create a Decoder with map type that preserves order
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var inventory Inventory
	err = decoder.Decode(&inventory)
	common.FailError(err, "error unmarshalling inventory: %v", filePath)

	hosts = nil // Clear the hosts slice to start fresh
	// Process "services" hosts
	for name, host := range inventory.Services.Hosts {
		common.Info("Found service host %s with IP %s", name, host.AnsibleHost)
		hosts = append(hosts, Host{
			Name:        name,
			AnsibleHost: host.AnsibleHost,
		})
	}

	return nil
}

// handleSSHKeysAndServiceChecks handles SSH key setup and service checks
func handleSSHKeysAndServiceChecks() {
	common.Info("Checking started...")

	// Print the list of hosts
	for _, host := range hosts {
		common.Info("Host: %s, IP: %s\n", host.Name, host.AnsibleHost)
	}

	for _, host := range hosts {
		if !checkSSHKeyExported(host.Name) {
			common.Info("Setting up SSH key for host %s\n", host.Name)
			err := setupSSHKeyForHost("master", host.Name, host.AnsibleHost)
			common.FailError(err, "Error setting up SSH keys for host %s: %v", host.Name, err)
		} else {
			common.Ok("SSH key already exported to host %s\n", host.Name)
		}

		// Check if the service exists and if not, run the setup playbook
		repoExists, serviceExists := checkServiceExists(host.Name, map[string]string{"service": host.Name}) // Use AnsibleHost (IP)
		if !repoExists || !serviceExists {
			common.Info("Setting up service for host %s\n", host.Name)

			// Pass the service name as an extra variable to the playbook
			err := runAnsiblePlaybook(config.AnsiblePath+"/setup.yaml", host.Name, map[string]string{"service": host.Name}) // Use AnsibleHost (IP)
			if err != nil {
				common.Err("Error setting up service for host %s: %v", host.Name, err) // Log error instead of failing
			}
		}
	}
}

func main() {
	common.Head("--DEPLOYER STARTS--")
	godotenv.Load("../.env")
	godotenv.Load(".env")

	common.PYTHON_API_HOST = os.Getenv("PYTHON_API_HOST")

	// 1. Load configuration
	err := LoadConfig("config.json")
	common.FailError(err, "Error loading configuration: %v\n")

	// 1.2 Load inventory
	readInventory(config.AnsiblePath + "/inventory.yaml")

	// 2. Initialize GitHub client
	client := NewGitHubClient(config.GitHubToken)

	// 3. Setup SSH Keys & Check Service Readiness
	handleSSHKeysAndServiceChecks()

	// 4. Watch for changes and deploy
	watchForChanges(client)
}
