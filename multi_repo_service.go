package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MultiRepoService manages multiple Git repositories
type MultiRepoService struct {
	config       *Config
	scanner      *RepoScanner
	repositories []*RepositoryInfo
	services     map[string]*GitAirService
	logger       *logrus.Logger
	mutex        sync.RWMutex
}

// MultiRepoConfig extends the base config for multi-repo support
type MultiRepoConfig struct {
	*Config
	ScanPaths    []string `yaml:"scan_paths"`
	MaxRepos     int      `yaml:"max_repos"`
	ScanInterval time.Duration `yaml:"scan_interval"`
}

// NewMultiRepoService creates a new multi-repository service
func NewMultiRepoService(config *MultiRepoConfig) (*MultiRepoService, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	scanner := NewRepoScanner(config.Config)

	return &MultiRepoService{
		config:   config.Config,
		scanner:  scanner,
		services: make(map[string]*GitAirService),
		logger:   logger,
	}, nil
}

// Start begins the multi-repository monitoring
func (mrs *MultiRepoService) Start(ctx context.Context) error {
	mrs.logger.Info("🏠 Git Air Multi-Repository Service Starting...")
	mrs.logger.Info("💻 Perfect for development servers - manages ALL your projects at once!")
	mrs.logger.Info("🔍 Will scan for Git repositories and auto-sync to ALL remotes")

	// Initial repository scan
	if err := mrs.discoverRepositories(); err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	// Start services for discovered repositories
	if err := mrs.startRepositoryServices(ctx); err != nil {
		return fmt.Errorf("failed to start repository services: %w", err)
	}

	// Create ticker for periodic repository discovery
	scanTicker := time.NewTicker(mrs.getScanInterval())
	defer scanTicker.Stop()

	mrs.logger.Info("Multi-repository service started successfully")

	for {
		select {
		case <-ctx.Done():
			mrs.logger.Info("Shutting down multi-repository service...")
			mrs.stopAllServices()
			return nil

		case <-scanTicker.C:
			mrs.logger.Debug("Performing periodic repository scan...")
			if err := mrs.periodicRepositoryScan(ctx); err != nil {
				mrs.logger.Errorf("Periodic repository scan failed: %v", err)
			}
		}
	}
}

// discoverRepositories discovers Git repositories in configured paths
func (mrs *MultiRepoService) discoverRepositories() error {
	scanPaths := mrs.getScanPaths()
	mrs.logger.Infof("Scanning for repositories in: %v", scanPaths)

	repositories, err := mrs.scanner.ScanForRepositories(scanPaths)
	if err != nil {
		return err
	}

	// Filter active repositories
	activeRepos := mrs.scanner.FilterActiveRepositories(repositories)

	mrs.mutex.Lock()
	mrs.repositories = activeRepos
	mrs.mutex.Unlock()

	// Print discovery report
	mrs.scanner.PrintRepositoryReport(activeRepos)

	return nil
}

// startRepositoryServices starts Git Air services for each repository
func (mrs *MultiRepoService) startRepositoryServices(ctx context.Context) error {
	mrs.mutex.RLock()
	repositories := mrs.repositories
	mrs.mutex.RUnlock()

	for _, repo := range repositories {
		if err := mrs.startServiceForRepository(ctx, repo); err != nil {
			mrs.logger.Errorf("Failed to start service for %s: %v", repo.Path, err)
			continue
		}
	}

	mrs.logger.Infof("Started services for %d repositories", len(mrs.services))
	return nil
}

// startServiceForRepository starts a Git Air service for a specific repository
func (mrs *MultiRepoService) startServiceForRepository(ctx context.Context, repo *RepositoryInfo) error {
	mrs.mutex.Lock()
	defer mrs.mutex.Unlock()

	// Check if service already exists
	if _, exists := mrs.services[repo.Path]; exists {
		mrs.logger.Debugf("Service already exists for %s", repo.Path)
		return nil
	}

	// Create repository-specific config
	repoConfig := *mrs.config
	repoConfig.CommitMessage = fmt.Sprintf("auto commit - %s", repo.Name)

	// Create and start service
	service, err := NewGitAirService(&repoConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Start service in background
	go func(repoPath string, svc *GitAirService) {
		mrs.logger.Infof("Starting Git Air service for: %s", repoPath)
		
		// Change to repository directory before starting
		originalDir := getCurrentDir()
		defer func() {
			if err := changeDir(originalDir); err != nil {
				mrs.logger.Errorf("Failed to restore directory: %v", err)
			}
		}()

		if err := changeDir(repoPath); err != nil {
			mrs.logger.Errorf("Failed to change to repository directory %s: %v", repoPath, err)
			return
		}

		if err := svc.Start(ctx); err != nil {
			mrs.logger.Errorf("Service failed for %s: %v", repoPath, err)
		}
	}(repo.Path, service)

	mrs.services[repo.Path] = service
	mrs.logger.Infof("✅ Started service for repository: %s", repo.Path)

	return nil
}

// periodicRepositoryScan performs periodic repository discovery
func (mrs *MultiRepoService) periodicRepositoryScan(ctx context.Context) error {
	// Discover new repositories
	if err := mrs.discoverRepositories(); err != nil {
		return err
	}

	// Start services for new repositories
	mrs.mutex.RLock()
	repositories := mrs.repositories
	mrs.mutex.RUnlock()

	for _, repo := range repositories {
		mrs.mutex.RLock()
		_, exists := mrs.services[repo.Path]
		mrs.mutex.RUnlock()

		if !exists {
			mrs.logger.Infof("New repository discovered: %s", repo.Path)
			if err := mrs.startServiceForRepository(ctx, repo); err != nil {
				mrs.logger.Errorf("Failed to start service for new repository %s: %v", repo.Path, err)
			}
		}
	}

	return nil
}

// stopAllServices stops all running repository services
func (mrs *MultiRepoService) stopAllServices() {
	mrs.mutex.Lock()
	defer mrs.mutex.Unlock()

	mrs.logger.Infof("Stopping %d repository services...", len(mrs.services))

	for repoPath := range mrs.services {
		mrs.logger.Infof("Stopping service for: %s", repoPath)
		// Services will stop when context is cancelled
	}

	mrs.services = make(map[string]*GitAirService)
	mrs.logger.Info("All repository services stopped")
}

// GetRepositoryStatus returns status information for all managed repositories
func (mrs *MultiRepoService) GetRepositoryStatus() map[string]interface{} {
	mrs.mutex.RLock()
	defer mrs.mutex.RUnlock()

	status := make(map[string]interface{})
	status["total_repositories"] = len(mrs.repositories)
	status["active_services"] = len(mrs.services)

	repoDetails := make([]map[string]interface{}, 0, len(mrs.repositories))
	for _, repo := range mrs.repositories {
		detail := map[string]interface{}{
			"name":         repo.Name,
			"path":         repo.Path,
			"remotes":      len(repo.Remotes),
			"service_active": mrs.services[repo.Path] != nil,
		}
		repoDetails = append(repoDetails, detail)
	}
	status["repositories"] = repoDetails

	return status
}

// getScanPaths returns the configured scan paths or defaults
func (mrs *MultiRepoService) getScanPaths() []string {
	// This could be extended to read from config
	return []string{"."} // Default to current directory
}

// getScanInterval returns the configured scan interval or default
func (mrs *MultiRepoService) getScanInterval() time.Duration {
	// Default to 5 minutes for repository discovery
	return 5 * time.Minute
}

// Helper functions
func changeDir(path string) error {
	return os.Chdir(path)
}