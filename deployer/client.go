package main

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/nesiler/cestx/common"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *github.Client
}

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

func (client *GitHubClient) PullLatest(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := cmd.CombinedOutput()
	common.FailError(err, "output: %s", err, string(output))
	return nil
}

func (c *GitHubClient) GetChangedDirs(repoPath, latestCommit string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--name-only", latestCommit+"^!")
	output, err := cmd.Output()
	common.FailError(err, "")

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

// watchForChanges watches for new commits and triggers deployments
func watchForChanges(client *GitHubClient) {
	var latestCommit string

	for {
		// Get the latest commit from GitHub
		commit, err := client.GetLatestCommit(config.RepoOwner, config.RepoName)
		common.FailError(err, "Error getting latest commit: %v\n")

		if commit != latestCommit {
			common.Ok("New commit detected: %s", commit)
			common.SendMessageToTelegram("New commit detected: " + commit)

			// Pull the latest changes from the repository
			common.FailError(client.PullLatest(config.RepoPath), "Error pulling latest changes")
			common.Ok("Pulled latest changes")

			// Get the list of changed directories in the repository
			changedDirs, err := client.GetChangedDirs(config.RepoPath, commit)
			common.FailError(err, "Error getting changed directories")

			common.Info("Changed directories: %v", changedDirs)

			// For each changed directory, deploy the corresponding service
			for _, dir := range changedDirs {
				common.Info("Trying to deploy: %s", dir)
				if err := Deploy(dir); err != nil {
					common.Warn("%v: %s", err, dir)
				} else {
					common.Ok("Successfully deployed: %s", dir)
					common.SendMessageToTelegram("Successfully deployed: " + dir)
				}
			}
			latestCommit = commit // Update latest commit hash
		}

		time.Sleep(time.Second * time.Duration(config.CheckInterval))
	}
}
