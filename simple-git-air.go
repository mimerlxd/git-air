package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	fmt.Println("🚀 Simple Git Air - Auto sync all Git repos")
	fmt.Println("📡 Inter-project communication via Git synchronization")
	
	// Find all git repos in current directory and subdirs
	repos, err := findGitRepos(".")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found %d Git repositories\n", len(repos))
	for _, repo := range repos {
		fmt.Printf("  📁 %s\n", repo)
	}
	
	// Main loop - check every 30 seconds for changes, pull every minute
	lastPull := time.Now()
	for {
		// Auto commit and push changes
		for _, repo := range repos {
			processRepo(repo)
		}
		
		// Pull from all repos every minute for inter-project communication
		if time.Since(lastPull) >= time.Minute {
			fmt.Println("\n📡 Checking for inter-project updates...")
			for _, repo := range repos {
				pullUpdates(repo)
			}
			lastPull = time.Now()
		}
		
		time.Sleep(30 * time.Second)
	}
}

// findGitRepos finds all .git directories
func findGitRepos(root string) ([]string, error) {
	var repos []string
	
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Skip some common dirs
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}
		
		// Found a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repos = append(repos, repoPath)
			return filepath.SkipDir // Don't go into .git
		}
		
		return nil
	})
	
	return repos, err
}

// processRepo handles one git repository
func processRepo(repoPath string) {
	// Change to repo directory
	oldDir, _ := os.Getwd()
	os.Chdir(repoPath)
	defer os.Chdir(oldDir)
	
	// Check if there are changes
	if !hasChanges() {
		return // No changes to commit
	}
	
	fmt.Printf("📝 %s: Auto committing changes...\n", filepath.Base(repoPath))
	
	// Auto commit
	runGit("add", ".")
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	runGit("commit", "-m", "auto commit - "+timestamp)
	
	// Push to all remotes immediately
	pushToAllRemotes()
}

// pullUpdates pulls from remotes for inter-project communication
func pullUpdates(repoPath string) {
	// Change to repo directory
	oldDir, _ := os.Getwd()
	os.Chdir(repoPath)
	defer os.Chdir(oldDir)
	
	pullFromRemotes()
}

// hasChanges checks if repo has uncommitted changes
func hasChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// pushToAllRemotes pushes to all configured remotes
func pushToAllRemotes() {
	remotes := getRemotes()
	if len(remotes) == 0 {
		return
	}
	
	branch := getCurrentBranch()
	for _, remote := range remotes {
		fmt.Printf("  🚀 Push to %s\n", remote)
		runGit("push", remote, branch)
	}
}

// pullFromRemotes pulls from remotes for inter-project communication
func pullFromRemotes() {
	remotes := getRemotes()
	if len(remotes) == 0 {
		return
	}
	
	branch := getCurrentBranch()
	repoName := filepath.Base(getCurrentDir())
	
	// Try to pull from each remote
	for _, remote := range remotes {
		fmt.Printf("  📥 %s: Checking %s for updates\n", repoName, remote)
		runGit("fetch", remote)
		
		// Check if there are remote changes
		if hasRemoteChanges(remote, branch) {
			fmt.Printf("  📡 %s: Pulling inter-project updates from %s\n", repoName, remote)
			runGit("pull", remote, branch)
		}
	}
}

// getRemotes returns list of remote names
func getRemotes() []string {
	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	
	remotes := strings.Fields(string(output))
	return remotes
}

// getCurrentBranch returns current branch name
func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "main" // fallback
	}
	return strings.TrimSpace(string(output))
}

// runGit runs a git command and returns success
func runGit(args ...string) bool {
	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

// hasRemoteChanges checks if remote has changes
func hasRemoteChanges(remote, branch string) bool {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	localOut, err := cmd.Output()
	if err != nil {
		return false
	}
	
	cmd = exec.Command("git", "rev-parse", remote+"/"+branch)
	remoteOut, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return string(localOut) != string(remoteOut)
}

// getCurrentDir returns current directory
func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}