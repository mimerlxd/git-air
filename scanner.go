package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// RepoScanner handles discovery and management of multiple Git repositories
type RepoScanner struct {
	config *Config
	logger *logrus.Logger
}

// RepositoryInfo contains information about a discovered Git repository
type RepositoryInfo struct {
	Path     string
	Name     string
	Remotes  []RemoteInfo
	GitRepo  *GitRepository
	IsActive bool
}

// RemoteInfo contains information about a Git remote
type RemoteInfo struct {
	Name string
	URL  string
}

// NewRepoScanner creates a new repository scanner
func NewRepoScanner(config *Config) *RepoScanner {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return &RepoScanner{
		config: config,
		logger: logger,
	}
}

// ScanForRepositories discovers Git repositories in the specified paths
func (rs *RepoScanner) ScanForRepositories(scanPaths []string) ([]*RepositoryInfo, error) {
	var repositories []*RepositoryInfo

	for _, scanPath := range scanPaths {
		repos, err := rs.scanPath(scanPath)
		if err != nil {
			rs.logger.Errorf("Failed to scan path %s: %v", scanPath, err)
			continue
		}
		repositories = append(repositories, repos...)
	}

	rs.logger.Infof("Discovered %d Git repositories", len(repositories))
	return repositories, nil
}

// scanPath recursively scans a path for Git repositories
func (rs *RepoScanner) scanPath(scanPath string) ([]*RepositoryInfo, error) {
	var repositories []*RepositoryInfo

	rs.logger.Debugf("Starting scan of path: %s", scanPath)

	err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			rs.logger.Debugf("Walk error at %s: %v", path, err)
			return err
		}

		// Skip excluded paths
		for _, exclude := range rs.config.ExcludePaths {
			if matched, _ := filepath.Match(exclude, filepath.Base(path)); matched {
				rs.logger.Debugf("Skipping excluded path: %s (matches %s)", path, exclude)
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if this is a Git repository
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			rs.logger.Debugf("Found .git directory at %s, analyzing repository at %s", path, repoPath)
			repo, err := rs.analyzeRepository(repoPath)
			if err != nil {
				rs.logger.Warnf("Failed to analyze repository at %s: %v", repoPath, err)
				return nil
			}

			repositories = append(repositories, repo)
			rs.logger.Infof("Found Git repository: %s", repoPath)

			// Skip scanning inside .git directory
			return filepath.SkipDir
		}

		return nil
	})

	rs.logger.Debugf("Scan completed. Found %d repositories in %s", len(repositories), scanPath)
	return repositories, err
}

// analyzeRepository analyzes a Git repository and extracts information
func (rs *RepoScanner) analyzeRepository(repoPath string) (*RepositoryInfo, error) {
	gitRepo, err := NewGitRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Get repository name
	repoName := filepath.Base(repoPath)

	// Get remotes
	remotes, err := rs.getRemotes(gitRepo)
	if err != nil {
		rs.logger.Warnf("Failed to get remotes for %s: %v", repoPath, err)
		remotes = []RemoteInfo{} // Continue with empty remotes
	}

	return &RepositoryInfo{
		Path:     repoPath,
		Name:     repoName,
		Remotes:  remotes,
		GitRepo:  gitRepo,
		IsActive: true,
	}, nil
}

// getRemotes retrieves all remotes for a Git repository
func (rs *RepoScanner) getRemotes(gitRepo *GitRepository) ([]RemoteInfo, error) {
	remotes, err := gitRepo.GetRemotes()
	if err != nil {
		return nil, err
	}

	var remoteInfos []RemoteInfo
	for name, url := range remotes {
		remoteInfos = append(remoteInfos, RemoteInfo{
			Name: name,
			URL:  url,
		})
	}

	return remoteInfos, nil
}

// FilterActiveRepositories filters repositories based on activity criteria
func (rs *RepoScanner) FilterActiveRepositories(repositories []*RepositoryInfo) []*RepositoryInfo {
	var activeRepos []*RepositoryInfo

	for _, repo := range repositories {
		if rs.isRepositoryActive(repo) {
			activeRepos = append(activeRepos, repo)
		} else {
			rs.logger.Debugf("Repository %s marked as inactive", repo.Path)
		}
	}

	rs.logger.Infof("Filtered to %d active repositories", len(activeRepos))
	return activeRepos
}

// isRepositoryActive determines if a repository should be actively monitored
func (rs *RepoScanner) isRepositoryActive(repo *RepositoryInfo) bool {
	// Repository is active if:
	// 1. It's not in an excluded path
	// 2. It has at least one remote (OR we allow repos without remotes for local work)

	// Check if repository path is excluded
	for _, exclude := range rs.config.ExcludePaths {
		if strings.Contains(repo.Path, exclude) {
			rs.logger.Debugf("Repository %s matches exclude pattern %s", repo.Path, exclude)
			return false
		}
	}

	// Allow repositories even without remotes (for local development)
	if len(repo.Remotes) == 0 {
		rs.logger.Debugf("Repository %s has no remotes, but keeping active for local work", repo.Path)
		return true // Changed: Keep repos without remotes active
	}

	return true
}

// GroupRepositoriesByRemote groups repositories by their remote URLs
func (rs *RepoScanner) GroupRepositoriesByRemote(repositories []*RepositoryInfo) map[string][]*RepositoryInfo {
	remoteGroups := make(map[string][]*RepositoryInfo)

	for _, repo := range repositories {
		for _, remote := range repo.Remotes {
			normalizedURL := rs.normalizeRemoteURL(remote.URL)
			remoteGroups[normalizedURL] = append(remoteGroups[normalizedURL], repo)
		}
	}

	rs.logger.Infof("Grouped repositories into %d remote groups", len(remoteGroups))
	return remoteGroups
}

// normalizeRemoteURL normalizes remote URLs for grouping
func (rs *RepoScanner) normalizeRemoteURL(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Convert SSH to HTTPS format for comparison
	if strings.HasPrefix(url, "git@") {
		url = strings.Replace(url, "git@", "https://", 1)
		url = strings.Replace(url, ":", "/", 1)
	}

	return strings.ToLower(url)
}

// PrintRepositoryReport prints a detailed report of discovered repositories
func (rs *RepoScanner) PrintRepositoryReport(repositories []*RepositoryInfo) {
	rs.logger.Info("🔍 Repository Discovery Report")
	rs.logger.Info("═══════════════════════════════")

	for i, repo := range repositories {
		rs.logger.Infof("📁 Repository %d: %s", i+1, repo.Name)
		rs.logger.Infof("   Path: %s", repo.Path)
		rs.logger.Infof("   Active: %v", repo.IsActive)
		rs.logger.Infof("   Remotes (%d):", len(repo.Remotes))

		for _, remote := range repo.Remotes {
			rs.logger.Infof("     - %s: %s", remote.Name, remote.URL)
		}

		rs.logger.Info("   ───────────────────────")
	}

	// Group statistics
	remoteGroups := rs.GroupRepositoriesByRemote(repositories)
	rs.logger.Infof("📊 Summary: %d repositories, %d unique remotes", len(repositories), len(remoteGroups))
}