# Git Air 🚀

A fully automatic Git daemon service for Linux and macOS that handles auto-commits, auto-pushes, and auto-pulls.

## Features

- **Auto Commit**: Automatically commits all changes with "auto commit" message
- **Auto Push**: Automatically pushes commits to remote repository  
- **Auto Pull**: Checks for remote changes every minute and pulls them
- **Multi-repo Support**: Works with regular repositories and monorepos
- **Cross Platform**: Supports Linux and macOS
- **Configurable**: YAML configuration file support
- **Daemon Mode**: Runs as background service
- **File System Monitoring**: Real-time file change detection

## Quick Start - Dev Server Setup

### Normal Case: Dev Server in HOME Directory
```bash
# 1. Place git-air in your HOME directory on dev server
cp git-air ~/
cd ~

# 2. Start Git Air - it will automatically discover and manage ALL projects
./git-air
```

**This will:**
- 🔍 **Scan your entire HOME directory** and all subdirectories
- 📁 **Discover ALL Git repositories** in your projects
- 🤖 **Auto-commit changes** in every repository  
- 🚀 **Push to ALL remotes** for each repository
- 🔄 **Keep everything synchronized** automatically

### Installation

### Prerequisites
- Go 1.21 or higher
- Git installed and configured

### Build from Source
```bash
git clone <repository-url>
cd git-air
go mod tidy
go build -o git-air
```

## Usage

### Basic Usage
```bash
# Run in current directory (scans current dir + all subdirectories for Git repos)
./git-air

# Run in specific directory and scan all subdirectories
./git-air -dir /path/to/your/projects

# Single repository mode (only monitor one specific Git repo)
./git-air -dir /path/to/single/repo -multi=false

# Custom scan paths (multiple directories)
./git-air -scan "/path/to/projects1,/path/to/projects2"

# Run with custom intervals
./git-air -watch 10s -pull 2m

# Run with custom log level
./git-air -log debug
```

### Default Behavior
**By default, `./git-air` will:**
- 🔍 **Scan current directory and ALL subdirectories** for Git repositories
- 📝 **Auto-commit** changes in ALL discovered repositories
- 🚀 **Push to ALL remotes** configured for each repository
- 📥 **Pull from remotes** to keep repositories synchronized
- 🔄 **Continuously monitor** for new repositories and changes

### Configuration File
Create a `git-air.yaml` file in your project directory:

```yaml
watch_interval: 30s
pull_interval: 1m
auto_commit: true
auto_push: true
auto_pull: true
commit_message: "auto commit"
exclude_paths:
  - ".git"
  - "node_modules"
  - "*.log"
log_level: "info"
```

### Command Line Options
- `-config`: Path to configuration file (default: "git-air.yaml")
- `-dir`: Directory to monitor (default: current directory)
- `-watch`: Interval between file system checks (default: 30s)
- `-pull`: Interval between remote checks (default: 1m)
- `-log`: Log level - debug, info, warn, error (default: "info")
- `-daemon`: Run as daemon process

## How It Works

1. **File Monitoring**: Uses filesystem watchers to detect changes in real-time
2. **Auto Commit**: When changes are detected, automatically stages and commits them
3. **Auto Push**: After successful commits, pushes to the configured remote
4. **Auto Pull**: Periodically checks remote repository for changes and pulls them
5. **Conflict Handling**: Handles merge conflicts gracefully with logging

## Use Cases

### 💻 **Primary Use Case: Development Server**
- **Place git-air in ~/home** on your development server
- **Manages all your projects automatically** - no manual setup per project
- **Multi-remote support** - pushes to origin, backup, mirror, etc.
- **Perfect for teams** - everyone's changes stay synchronized
- **Never lose work** - continuous auto-commits serve as safety net

### 🔄 **Other Use Cases**

- **Development Workflow**: Never lose work with automatic commits
- **Team Collaboration**: Keep repositories synchronized automatically  
- **CI/CD Integration**: Ensure changes are always pushed to trigger builds
- **Backup Strategy**: Automatic commits serve as frequent backups
- **Inter-project Communication**: Auto-sync between related repositories

## Architecture

The service consists of several key components:

- **Main Service**: Orchestrates all operations and handles signals
- **Git Repository**: Wraps Git operations with error handling
- **File Watcher**: Monitors filesystem changes using fsnotify
- **Configuration**: YAML-based configuration management
- **Logging**: Structured logging with configurable levels

## Security Considerations

- Git Air only operates on the local repository
- Uses standard Git commands - no direct repository manipulation
- Respects Git configuration (credentials, remotes, etc.)
- Excludes sensitive files through configuration

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues and feature requests, please use the GitHub issue tracker.