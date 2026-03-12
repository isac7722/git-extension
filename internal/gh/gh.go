package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes a gh CLI command and returns its stdout.
func Run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("gh %s: %s", strings.Join(args, " "), errMsg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// CurrentUsername returns the authenticated GitHub username.
func CurrentUsername() (string, error) {
	return Run("api", "user", "--jq", ".login")
}

// RepoCollaborators returns the list of collaborator usernames for the current repo.
func RepoCollaborators() ([]string, error) {
	nameWithOwner, err := Run("repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	if err != nil {
		return nil, fmt.Errorf("failed to get repo info: %w", err)
	}

	output, err := Run("api", fmt.Sprintf("repos/%s/collaborators", nameWithOwner), "--jq", ".[].login")
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	if output == "" {
		return nil, nil
	}

	var logins []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			logins = append(logins, line)
		}
	}
	return logins, nil
}

// RepoCollaboratorsWithNames returns collaborators with their display names.
type Collaborator struct {
	Login string
	Name  string
}

// RepoCollaboratorsDetailed returns collaborators with login and name.
func RepoCollaboratorsDetailed(nameWithOwner string) ([]Collaborator, error) {
	output, err := Run("api", fmt.Sprintf("repos/%s/collaborators", nameWithOwner))
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	var raw []struct {
		Login string `json:"login"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse collaborators: %w", err)
	}

	var collaborators []Collaborator
	for _, r := range raw {
		collaborators = append(collaborators, Collaborator{Login: r.Login, Name: r.Name})
	}
	return collaborators, nil
}
