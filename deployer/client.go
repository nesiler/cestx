package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/nesiler/cestx/common"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *github.Client
}

var client *GitHubClient

func NewGitHubClient(token string) *GitHubClient {
	common.Info("Creating GitHub client")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	common.Ok("GitHub client created: %v", client.BaseURL.String())
	return &GitHubClient{client: client}
}

func (c *GitHubClient) GetLatestCommit(owner, repo string) (string, error) {
	ctx := context.Background()
	commits, _, err := c.client.Repositories.ListCommits(ctx, owner, repo, nil)
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", nil
	}
	return *commits[0].SHA, nil
}

func (c *GitHubClient) GetChangedDirs(repoPath, latestCommit string) ([]string, error) {
	common.Warn("Getting changed directories between %s and %s", lastKnownCommit, latestCommit)
	common.Out("Checking for changes in %s", repoPath)
	// currentCommit, err := getCurrentCommit(repoPath)
	// if err != nil {
	// 	return nil, err
	// }

	cmd := exec.Command("git", "-C", repoPath, "diff", "--name-only", lastKnownCommit, latestCommit)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	changedFiles := strings.Split(out.String(), "\n")
	changedDirs := make(map[string]bool)
	for _, file := range changedFiles {
		if dir := filepath.Dir(file); dir != "." {
			changedDirs[dir] = true
		}
	}

	var dirs []string
	for dir := range changedDirs {
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

func getCurrentCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// watchForChanges watches for new commits and triggers deployments
func watchForChanges() {
	// Ideally, load latest known commit from a more persistent storage
	deployerChanged := false

	latestCommit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
	if err != nil {
		common.Err("Failed to fetch the latest commit: %v", err)
		return
	}

	if latestCommit != lastKnownCommit {
		common.Info("New commit detected: %s", latestCommit)
		common.SendMessageToTelegram("**DEPLOYER** ::: New commit detected: " + latestCommit)
		changedDirs, err := client.GetChangedDirs(config.RepoPath, latestCommit)
		if err != nil {
			common.Err("Failed to get changed directories: %v", err)
			return
		}

		for _, dir := range changedDirs {
			if dir != "deployer" {
				common.Info("Deploying changes for directory: %s", dir)
				err := Deploy(dir)
				if err != nil {
					common.Err("Failed to deploy service %s: %v", dir, err)
				}
			} else if dir == "deployer" {
				deployerChanged = true
			}
		}
		if deployerChanged {
			starterService()
		}
		lastKnownCommit = latestCommit
	}
}

func starterService() {
	common.Info("Deployer changes detected; pulling updates and restarting...")

	// Restart the starter.service to apply changes
	cmd := exec.Command("systemctl", "restart", "starter.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		common.Err("Failed to restart starter.service: %v", err)
	}

	// it should already be restarted, but just in case, sleep for a few seconds
	time.Sleep(5 * time.Second)

}
