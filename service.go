package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// GitAirService manages the Git Air daemon functionality
type GitAirService struct {
	config  *Config
	watcher *fsnotify.Watcher
	gitRepo *GitRepository
	logger  *logrus.Logger
}

// NewGitAirService creates a new Git Air service instance
func NewGitAirService(config *Config) (*GitAirService, error) {
	// Initialize file system watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Initialize Git repository
	gitRepo, err := NewGitRepository(".")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return &GitAirService{
		config:  config,
		watcher: watcher,
		gitRepo: gitRepo,
		logger:  logger,
	}, nil
}

// Start begins the Git Air daemon
func (s *GitAirService) Start(ctx context.Context) error {
	s.logger.Info("🚀 Git Air daemon starting...")

	// Add current directory to watcher
	if err := s.addWatchPaths("."); err != nil {
		return fmt.Errorf("failed to setup file watching: %w", err)
	}

	// Create tickers for periodic operations
	watchTicker := time.NewTicker(s.config.WatchInterval)
	pullTicker := time.NewTicker(s.config.PullInterval)

	defer func() {
		watchTicker.Stop()
		pullTicker.Stop()
		s.watcher.Close()
	}()

	s.logger.Infof("Monitoring directory: %s", getCurrentDir())
	s.logger.Infof("Watch interval: %v", s.config.WatchInterval)
	s.logger.Infof("Pull interval: %v", s.config.PullInterval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Shutting down Git Air daemon...")
			return nil

		case event, ok := <-s.watcher.Events:
			if !ok {
				return fmt.Errorf("file watcher channel closed")
			}
			s.handleFileEvent(event)

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return fmt.Errorf("file watcher error channel closed")
			}
			s.logger.Errorf("File watcher error: %v", err)

		case <-watchTicker.C:
			if s.config.AutoCommit {
				s.performAutoCommit()
			}

		case <-pullTicker.C:
			if s.config.AutoPull {
				s.performAutoPull()
			}
		}
	}
}

// addWatchPaths adds directories to the file system watcher
func (s *GitAirService) addWatchPaths(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded paths
		for _, exclude := range s.config.ExcludePaths {
			if matched, _ := filepath.Match(exclude, filepath.Base(path)); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Only watch directories
		if info.IsDir() {
			s.logger.Debugf("Watching directory: %s", path)
			return s.watcher.Add(path)
		}

		return nil
	})
}

// handleFileEvent processes file system events
func (s *GitAirService) handleFileEvent(event fsnotify.Event) {
	s.logger.Debugf("File event: %s %s", event.Op, event.Name)

	// Skip excluded files
	for _, exclude := range s.config.ExcludePaths {
		if matched, _ := filepath.Match(exclude, filepath.Base(event.Name)); matched {
			return
		}
	}

	// Trigger auto commit after a short delay to batch changes
	if s.config.AutoCommit {
		go func() {
			time.Sleep(2 * time.Second)
			s.performAutoCommit()
		}()
	}
}

// performAutoCommit executes automatic git commit
func (s *GitAirService) performAutoCommit() {
	if !s.gitRepo.HasChanges() {
		return
	}

	s.logger.Info("📝 Changes detected, performing auto commit...")

	// Add all changes
	if err := s.gitRepo.AddAll(); err != nil {
		s.logger.Errorf("Failed to add changes: %v", err)
		return
	}

	// Create commit message with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	commitMessage := fmt.Sprintf("%s - %s", s.config.CommitMessage, timestamp)

	// Commit changes
	if err := s.gitRepo.Commit(commitMessage); err != nil {
		s.logger.Errorf("Failed to commit changes: %v", err)
		return
	}

	s.logger.Info("✅ Auto commit successful")

	// Auto push if enabled
	if s.config.AutoPush {
		s.performAutoPush()
	}
}

// performAutoPush executes automatic git push
func (s *GitAirService) performAutoPush() {
	s.logger.Info("🔄 Performing auto push...")

	if err := s.gitRepo.Push(); err != nil {
		s.logger.Errorf("Failed to push changes: %v", err)
		return
	}

	s.logger.Info("✅ Auto push successful")
}

// performAutoPull executes automatic git pull
func (s *GitAirService) performAutoPull() {
	s.logger.Debug("🔄 Checking for remote changes...")

	hasRemoteChanges, err := s.gitRepo.HasRemoteChanges()
	if err != nil {
		s.logger.Errorf("Failed to check remote changes: %v", err)
		return
	}

	if !hasRemoteChanges {
		return
	}

	s.logger.Info("📥 Remote changes detected, performing auto pull...")

	if err := s.gitRepo.Pull(); err != nil {
		s.logger.Errorf("Failed to pull changes: %v", err)
		return
	}

	s.logger.Info("✅ Auto pull successful")
}

// getCurrentDir returns the current working directory
func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}