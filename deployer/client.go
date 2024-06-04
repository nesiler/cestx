package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	if latestCommit == "" {
		return nil, fmt.Errorf("no parent commit found for the initial commit")
	}

	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "-C", repoPath, "diff", "--name-only", latestCommit+"^!")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running git diff: %s, %s", stderr.String(), err)
	}

	output := out.Bytes()
	common.Warn("Changed dirs output: %s", output)

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	dirSet := make(map[string]struct{})
	for _, file := range changedFiles {
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			dirSet[parts[0]] = struct{}{}
		}
	}

	var dirs []string
	for dir := range dirSet {
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
	latestCommit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
	if err != nil {
		common.Err("Error getting latest commit: %v", err)
	}

	lastDeployedCommit, err := getCurrentCommit(config.RepoPath)
	if err != nil {
		common.Err("Error reading last deployed commit: %v", err)
	}

	if latestCommit != lastDeployedCommit {
		common.Info("New commit detected: %s", latestCommit)

		changedDirs, err := client.GetChangedDirs(config.RepoPath, latestCommit)
		common.Out("Changed directories: %v", changedDirs)
		common.SendMessageToTelegram("**DEPLOYER** ::: New Commit :: Changed directories: " + strings.Join(changedDirs, ", "))
		if err != nil {
			common.Err("Error getting changed directories: %v", err)
		}

		for _, dir := range changedDirs {
			if dir != "deployer" {
				err = Deploy(dir)
				if err != nil {
					common.Err("Error deploying service %s: %v", dir, err)
				}
			} else {
				common.Info("Skipping deployer directory")
				continue
			}
		}
	}
	common.Info("No changes detected")
}
