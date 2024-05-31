package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	"github.com/nesiler/cestx/common"
	"gopkg.in/yaml.v2"
)

// Host represents a single host in the inventory.
type Host struct {
	AnsibleHost              string `yaml:"ansible_host"`
	AnsibleSSHPrivateKeyFile string `yaml:"ansible_ssh_private_key_file"`
}

// Inventory represents the structure of the inventory YAML file.
type Inventory struct {
	All struct {
		Hosts map[string]Host `yaml:"hosts"`
	} `yaml:"all"`
}

// readInventory reads the inventory file and returns a slice of ansible_host values.
func readInventory(filePath string) ([]string, error) {

	data, err := os.ReadFile(filePath)
	common.FailError(err, "error reading inventory file: %v", filePath)

	var inventory Inventory
	err = yaml.Unmarshal(data, &inventory)
	common.FailError(err, "error unmarshalling inventory: %v", filePath)

	hosts := []string{}
	for name, host := range inventory.All.Hosts {
		common.Info("Found host %s with IP %s\n", name, host.AnsibleHost)
		hosts = append(hosts, host.AnsibleHost)
	}

	return hosts, nil
}

func checkSSHKeyExported(hosts []string) bool {
	for _, host := range hosts {
		cmd := exec.Command("ansible", host, "-m", "ping", "--private-key", os.Getenv("HOME")+"/.ssh/master", "-u", "root")
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}

func Deploy(config *Config, serviceName string) error {
	cmd := exec.Command("ansible-playbook", "-i", config.AnsiblePath+"/inventory.yaml", config.AnsiblePath+"/deploy.yml", "-e", "service="+serviceName)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

func main() {
	godotenv.Load("../.env")
	inventoryPath := "./ansible/inventory.yaml"
	hosts, err := readInventory(inventoryPath)
	if err != nil {
		common.Fatal("Error reading inventory: %v", err)
	}

	if !checkSSHKeyExported(hosts) {
		if err := setupSSHKeysForHosts("master", "root", hosts); err != nil {
			common.Fatal("Error setting up SSH keys: %v", err)
		}
	} else {
		common.Ok("SSH keys already exported to all hosts\n")
	}

	// Load configuration
	config, err := LoadConfig("config.json")
	common.FailError(err, "Error loading configuration: %v\n", err)

	// Get the Python API host from dotenv
	common.PYTHON_API_HOST = os.Getenv("PYTHON_API_HOST")

	// Initialize GitHub client
	client := NewGitHubClient(config.GitHubToken)

	// Store the latest commit hash
	var latestCommit string

	for {
		commit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
		if err != nil {
			common.Warn("Error getting latest commit: %v\n", err)
			time.Sleep(time.Second * 10)
			continue
		}

		if commit != latestCommit {
			common.Ok("New commit detected: %s\n", commit)
			common.SendMessageToTelegram("New commit detected: " + commit)
			latestCommit = commit
			err := client.PullLatest(config.RepoPath)
			if err != nil {
				common.Warn("Error pulling latest changes: %v\n", err)
				continue
			}

			changedDirs, err := client.GetChangedDirs(config.RepoPath, latestCommit)
			if err != nil {
				common.Warn("Error getting changed directories: %v\n", err)
				continue
			}

			for _, dir := range changedDirs {
				err := Deploy(config, dir)
				if err != nil {
					common.Warn("Error deploying %s: %v\n", dir, err)
				} else {
					common.Head("Successfully deployed: %s\n", dir)
					common.SendMessageToTelegram("Successfully deployed: " + dir)
				}
			}
		}

		time.Sleep(time.Second * time.Duration(config.CheckInterval))
	}
}
