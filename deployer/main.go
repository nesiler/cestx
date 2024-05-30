package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/nesiler/cestx/common"
	"gopkg.in/ini.v1"
)

func readInventory(filePath string) ([]string, error) {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, common.Fatal("error loading inventory: %v", err)
	}

	hosts := []string{}
	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}
		for _, key := range section.Keys() {
			if ansibleHost := key.String(); ansibleHost != "" {
				hosts = append(hosts, ansibleHost)
			}
		}
	}

	return hosts, nil
}

func Deploy(config *Config, serviceName string) error {
	cmd := exec.Command("ansible-playbook", "-i", config.AnsiblePath+"/inventory.ini", config.AnsiblePath+"/deploy.yml", "-e", "service="+serviceName, "--private-key", os.Getenv("HOME")+"/.ssh/master")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

func main() {

	inventoryPath := "./ansible/inventory.ini"
	hosts, err := readInventory(inventoryPath)
	if err != nil {
		common.Fatal("Error reading inventory: %v", err)
	}

	if err := setupSSHKeysForHosts("master", "root", hosts); err != nil {
		common.Fatal("Error setting up SSH keys: %v", err)
	}

	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize GitHub client
	client := NewGitHubClient(config.GitHubToken)

	// Store the latest commit hash
	var latestCommit string

	for {
		commit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
		if err != nil {
			common.Err("Error getting latest commit: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		if commit != latestCommit {
			log.Printf("New commit detected: %s", commit)
			// Send message to telegram with common package
			latestCommit = commit
			err := client.PullLatest(config.RepoPath)
			if err != nil {
				log.Printf("Error pulling latest changes: %v", err)
				continue
			}

			changedDirs, err := client.GetChangedDirs(config.RepoPath, latestCommit)
			if err != nil {
				log.Printf("Error getting changed directories: %v", err)
				continue
			}

			for _, dir := range changedDirs {
				err := Deploy(config, dir)
				if err != nil {
					log.Printf("Error deploying %s: %v", dir, err)
				}
			}
		}

		time.Sleep(time.Second * time.Duration(config.CheckInterval))
	}
}
