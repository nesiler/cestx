package main

import (
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
	All struct {
		Hosts map[string]Host `yaml:"hosts"`
	} `yaml:"all"`
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

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	common.FailError(err, "error reading config file: %v")

	var config Config
	err = json.Unmarshal(data, &config)
	common.FailError(err, "error parsing config file: %v")

	return &config, nil
}

// readInventory reads the inventory file and returns a map of hostnames to ansible_host values.
func readInventory(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	common.FailError(err, "error reading inventory file: %v", filePath)

	var inventory Inventory
	err = yaml.Unmarshal(data, &inventory)
	common.FailError(err, "error unmarshalling inventory: %v", filePath)

	hosts := make(map[string]string)
	for name, host := range inventory.All.Hosts {
		common.Info("Found host %s with IP %s\n", name, host.AnsibleHost)
		hosts[name] = host.AnsibleHost
	}

	return hosts, nil
}

// handleSSHKeysAndServiceChecks handles SSH key setup and service checks
func handleSSHKeysAndServiceChecks(config *Config) {
	inventoryPath := config.AnsiblePath + "/inventory.yaml"
	hosts, err := readInventory(inventoryPath)
	common.FailError(err, "Error reading inventory: %v")

	for name, ip := range hosts {
		// Check if SSH key is exported and if not, export it
		if !checkSSHKeyExported(name) {
			common.Info("Setting up SSH key for host %s\n", name)
			common.FailError(setupSSHKeyForHost("master", name, ip), "Error setting up SSH keys for host %s: %v", name, err)
		} else {
			common.Ok("SSH key already exported to host %s\n", name)
		}

		// Check if the service exists and if not, run the setup playbook
		if !checkServiceExists(name, map[string]string{"service": name}) {
			common.Info("Setting up service for host %s\n", name)

			// Pass the service name as an extra variable to the playbook
			err := runAnsiblePlaybook(config.AnsiblePath+"/setup.yml", name, map[string]string{"service": name})
			if err != nil {
				common.Err("Error setting up service for host %s: %v", name, err) // Log error instead of failing
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
	config, err := LoadConfig("config.json")
	common.FailError(err, "Error loading configuration: %v\n")

	// 2. Initialize GitHub client
	client := NewGitHubClient(config.GitHubToken)

	// 3. Setup SSH Keys & Check Service Readiness
	go handleSSHKeysAndServiceChecks(config) // Run in a separate goroutine

	// 4. Watch for changes and deploy
	watchForChanges(config, client)
}
