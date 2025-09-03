package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitRepository represents a Git repository and provides Git operations
type GitRepository struct {
	path string
}

// NewGitRepository creates a new GitRepository instance
func NewGitRepository(path string) (*GitRepository, error) {
	repo := &GitRepository{path: path}

	// Verify this is a git repository
	if !repo.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	return repo, nil
}

// IsGitRepository checks if the current directory is a git repository
func (r *GitRepository) IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = r.path
	err := cmd.Run()
	return err == nil
}

// HasChanges checks if there are any uncommitted changes
func (r *GitRepository) HasChanges() bool {
	// Check for staged changes
	cmd := exec.Command("git", "diff-index", "--quiet", "--cached", "HEAD", "--")
	cmd.Dir = r.path
	if cmd.Run() != nil {
		return true
	}

	// Check for unstaged changes
	cmd = exec.Command("git", "diff-index", "--quiet", "HEAD", "--")
	cmd.Dir = r.path
	return cmd.Run() != nil
}

// AddAll stages all changes
func (r *GitRepository) AddAll() error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %s", string(output))
	}
	return nil
}

// Commit creates a new commit with the given message
func (r *GitRepository) Commit(message string) error {
	// Check if there are staged changes
	cmd := exec.Command("git", "diff-index", "--quiet", "--cached", "HEAD", "--")
	cmd.Dir = r.path
	if cmd.Run() == nil {
		return nil // No staged changes to commit
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %s", string(output))
	}
	return nil
}

// Push pushes commits to ALL remote repositories
func (r *GitRepository) Push() error {
	remotes, err := r.GetRemotes()
	if err != nil {
		return fmt.Errorf("failed to get remotes: %w", err)
	}

	if len(remotes) == 0 {
		return fmt.Errorf("no remote repository configured")
	}

	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Push to ALL remotes
	var errors []string
	for remoteName := range remotes {
		cmd := exec.Command("git", "push", remoteName, branch)
		cmd.Dir = r.path
		output, err := cmd.CombinedOutput()
		if err != nil {
			errors = append(errors, fmt.Sprintf("push to %s failed: %s", remoteName, string(output)))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("push failures: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Pull pulls changes from ALL remote repositories
func (r *GitRepository) Pull() error {
	remotes, err := r.GetRemotes()
	if err != nil {
		return fmt.Errorf("failed to get remotes: %w", err)
	}

	if len(remotes) == 0 {
		return fmt.Errorf("no remote repository configured")
	}

	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Try to pull from each remote (typically origin first)
	var lastError error
	for remoteName := range remotes {
		cmd := exec.Command("git", "pull", remoteName, branch)
		cmd.Dir = r.path
		output, err := cmd.CombinedOutput()
		if err != nil {
			lastError = fmt.Errorf("git pull from %s failed: %s", remoteName, string(output))
		} else {
			// Successfully pulled from this remote
			return nil
		}
	}

	// If we get here, all pulls failed
	if lastError != nil {
		return lastError
	}

	return nil
}

// HasRemoteChanges checks if there are changes in the remote repository
func (r *GitRepository) HasRemoteChanges() (bool, error) {
	// Fetch latest changes
	if err := r.Fetch(); err != nil {
		return false, err
	}

	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return false, err
	}

	// Compare local and remote commits
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.path
	localOutput, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get local commit: %w", err)
	}
	localCommit := strings.TrimSpace(string(localOutput))

	cmd = exec.Command("git", "rev-parse", fmt.Sprintf("origin/%s", branch))
	cmd.Dir = r.path
	remoteOutput, err := cmd.Output()
	if err != nil {
		// Remote branch might not exist
		return false, nil
	}
	remoteCommit := strings.TrimSpace(string(remoteOutput))

	return localCommit != remoteCommit, nil
}

// GetStatus returns the git status
func (r *GitRepository) GetStatus() (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}
	return string(output), nil
}

// Fetch fetches changes from the remote repository
func (r *GitRepository) Fetch() error {
	if !r.HasRemote() {
		return nil // No remote to fetch from
	}

	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %s", string(output))
	}
	return nil
}

// GetCurrentBranch returns the name of the current branch
func (r *GitRepository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// HasRemote checks if a remote repository is configured
func (r *GitRepository) HasRemote() bool {
	cmd := exec.Command("git", "remote")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "origin")
}

// GetRemotes returns all configured remotes
func (r *GitRepository) GetRemotes() (map[string]string, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git remote failed: %w", err)
	}

	remotes := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			remoteName := parts[0]
			remoteURL := parts[1]
			// Only take fetch URLs (ignore push URLs)
			if len(parts) < 3 || parts[2] == "(fetch)" {
				remotes[remoteName] = remoteURL
			}
		}
	}

	return remotes, nil
}

// PushToRemote pushes to a specific remote
func (r *GitRepository) PushToRemote(remoteName string) error {
	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	cmd := exec.Command("git", "push", remoteName, branch)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push to %s failed: %s", remoteName, string(output))
	}
	return nil
}

// PullFromRemote pulls from a specific remote
func (r *GitRepository) PullFromRemote(remoteName string) error {
	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	cmd := exec.Command("git", "pull", remoteName, branch)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull from %s failed: %s", remoteName, string(output))
	}
	return nil
}

// FetchFromRemote fetches from a specific remote
func (r *GitRepository) FetchFromRemote(remoteName string) error {
	cmd := exec.Command("git", "fetch", remoteName)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch from %s failed: %s", remoteName, string(output))
	}
	return nil
}

// HasRemoteChanges checks if there are changes in a specific remote
func (r *GitRepository) HasRemoteChangesForRemote(remoteName string) (bool, error) {
	// Fetch latest changes from specific remote
	if err := r.FetchFromRemote(remoteName); err != nil {
		return false, err
	}

	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return false, err
	}

	// Compare local and remote commits
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.path
	localOutput, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get local commit: %w", err)
	}
	localCommit := strings.TrimSpace(string(localOutput))

	cmd = exec.Command("git", "rev-parse", fmt.Sprintf("%s/%s", remoteName, branch))
	cmd.Dir = r.path
	remoteOutput, err := cmd.Output()
	if err != nil {
		// Remote branch might not exist
		return false, nil
	}
	remoteCommit := strings.TrimSpace(string(remoteOutput))

	return localCommit != remoteCommit, nil
}