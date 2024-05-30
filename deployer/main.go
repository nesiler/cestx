package main

import (
	"log"
	"os/exec"
	"time"

	"github.com/nesiler/cestx/common"
	"gopkg.in/ini.v1"
)

func readInventory(filePath string) ([]string, error) {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, common.Err("error reading inventory file: %v", err)
	}

	hosts := []string{}
	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}
		for _, key := range section.Keys() {
			if key.Name() == "ansible_host" {
				common.Info("Found host: %s\n", key.String())
				hosts = append(hosts, key.String())
			}
		}
	}

	return hosts, nil
}

func checkSSHKeyExported(hosts []string) bool {
	for _, host := range hosts {
		cmd := exec.Command("ansible", host, "-m", "ping")
		if err := cmd.Run(); err != nil {
			return false
		}
	}
	return true
}

func Deploy(config *Config, serviceName string) error {
	cmd := exec.Command("ansible-playbook", "-i", config.AnsiblePath+"/inventory.ini", config.AnsiblePath+"/deploy.yml", "-e", "service="+serviceName)
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

	if !checkSSHKeyExported(hosts) {
		if err := setupSSHKeysForHosts("master", "root", hosts); err != nil {
			common.Fatal("Error setting up SSH keys: %v", err)
		}
	} else {
		common.Ok("SSH keys already exported to all hosts\n")
	}

	// Load configuration
	config, err := LoadConfig("config.json")
	common.Err("Error loading configuration: %v\n", err)

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
