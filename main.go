package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultWatchInterval = 30 * time.Second
	defaultPullInterval  = 60 * time.Second
)

func main() {
	var (
		configPath    = flag.String("config", "git-air.yaml", "Path to configuration file")
		workDir       = flag.String("dir", ".", "Directory to monitor")
		watchInterval = flag.Duration("watch", defaultWatchInterval, "Interval between file system checks")
		pullInterval  = flag.Duration("pull", defaultPullInterval, "Interval between remote checks")
		logLevel      = flag.String("log", "info", "Log level (debug, info, warn, error)")
		multiRepo     = flag.Bool("multi", false, "Enable multi-repository mode")
		scanPaths     = flag.String("scan", ".", "Comma-separated paths to scan for repositories")
		_             = flag.Bool("daemon", false, "Run as daemon")
	)
	flag.Parse()

	// Setup logging
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Fatal("Invalid log level:", err)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Change to working directory
	if err := os.Chdir(*workDir); err != nil {
		logrus.Fatal("Failed to change directory:", err)
	}

	workingDir, _ := os.Getwd()
	logrus.Infof("🏠 Git Air starting in: %s", workingDir)
	logrus.Info("💻 Dev Server Mode: Will scan all subdirectories for Git repositories")
	logrus.Info("🚀 Multi-remote support: Pushes to ALL configured remotes")

	// Load configuration
	config, err := LoadConfig(*configPath)
	if err != nil {
		logrus.Warnf("Failed to load config file %s, using defaults: %v", *configPath, err)
		config = DefaultConfig()
	}

	// Override config with command line flags
	config.WatchInterval = *watchInterval
	config.PullInterval = *pullInterval
	config.MultiRepo = *multiRepo
	if *scanPaths != "." {
		config.ScanPaths = strings.Split(*scanPaths, ",")
	} else {
		// Default behavior: scan current directory and all subdirectories
		config.ScanPaths = []string{*workDir}
		config.MultiRepo = true // Enable multi-repo by default
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logrus.Infof("Received signal %s, shutting down...", sig)
		cancel()
	}()

	// Start appropriate mode
	if config.MultiRepo {
		logrus.Info("🔍 Starting Git Air in multi-repository mode...")
		if err := startMultiRepoMode(ctx, config); err != nil {
			logrus.Fatal("Multi-repo service failed:", err)
		}
	} else {
		logrus.Info("📁 Starting Git Air in single-repository mode...")
		if err := startSingleRepoMode(ctx, config, *workDir); err != nil {
			logrus.Fatal("Single-repo service failed:", err)
		}
	}

	logrus.Info("Git Air daemon stopped")
}

// startMultiRepoMode starts Git Air for multiple repositories
func startMultiRepoMode(ctx context.Context, config *Config) error {
	multiConfig := &MultiRepoConfig{
		Config:       config,
		ScanPaths:    config.ScanPaths,
		MaxRepos:     config.MaxRepos,
		ScanInterval: config.ScanInterval,
	}

	service, err := NewMultiRepoService(multiConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize multi-repo service: %w", err)
	}

	return service.Start(ctx)
}

// startSingleRepoMode starts Git Air for a single repository
func startSingleRepoMode(ctx context.Context, config *Config, workDir string) error {
	// Change to working directory
	if err := os.Chdir(workDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	workingDir, _ := os.Getwd()
	logrus.Infof("📁 Git Air monitoring: %s", workingDir)

	// Initialize Git Air service
	service, err := NewGitAirService(config)
	if err != nil {
		return fmt.Errorf("failed to initialize Git Air service: %w", err)
	}

	return service.Start(ctx)
}