#!/bin/bash

# Git Air Daemon - Continuous Auto Commit Service
# Monitors for changes and automatically commits with "auto commit"

# Configuration
WATCH_INTERVAL=30  # seconds between checks
LOG_FILE="git-air.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

# Signal handler for graceful shutdown
cleanup() {
    log "INFO" "${YELLOW}Git Air daemon shutting down...${NC}"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log "ERROR" "${RED}Not a git repository${NC}"
        exit 1
    fi
}

# Auto commit function
auto_commit() {
    # Check for changes
    if git diff-index --quiet HEAD -- 2>/dev/null; then
        return 0  # No changes
    fi

    log "INFO" "${BLUE}Changes detected, preparing auto commit...${NC}"

    # Add all changes
    git add . 2>/dev/null

    # Check if there are staged changes
    if git diff-index --quiet --cached HEAD -- 2>/dev/null; then
        return 0  # No staged changes
    fi

    # Commit with auto commit message
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    if git commit -m "auto commit - $timestamp" >/dev/null 2>&1; then
        log "INFO" "${GREEN}Auto commit successful${NC}"
        
        # Try to push if remote is configured
        if git remote | grep -q origin 2>/dev/null; then
            if git push origin HEAD >/dev/null 2>&1; then
                log "INFO" "${GREEN}Auto push successful${NC}"
            else
                log "WARN" "${YELLOW}Auto push failed${NC}"
            fi
        fi
        return 0
    else
        log "ERROR" "${RED}Auto commit failed${NC}"
        return 1
    fi
}

# Auto pull function
auto_pull() {
    if git remote | grep -q origin 2>/dev/null; then
        # Fetch latest changes
        if git fetch origin >/dev/null 2>&1; then
            # Check if there are remote changes
            local local_commit=$(git rev-parse HEAD)
            local remote_commit=$(git rev-parse origin/$(git branch --show-current) 2>/dev/null)
            
            if [ "$local_commit" != "$remote_commit" ] && [ -n "$remote_commit" ]; then
                log "INFO" "${BLUE}Remote changes detected, pulling...${NC}"
                if git pull origin $(git branch --show-current) >/dev/null 2>&1; then
                    log "INFO" "${GREEN}Auto pull successful${NC}"
                else
                    log "WARN" "${YELLOW}Auto pull failed (possible conflicts)${NC}"
                fi
            fi
        fi
    fi
}

# Main daemon loop
main() {
    check_git_repo
    
    log "INFO" "${GREEN}🚀 Git Air daemon started${NC}"
    log "INFO" "Monitoring directory: $(pwd)"
    log "INFO" "Watch interval: ${WATCH_INTERVAL}s"
    log "INFO" "Log file: $LOG_FILE"
    log "INFO" "Press Ctrl+C to stop"
    
    while true; do
        auto_commit
        auto_pull
        sleep "$WATCH_INTERVAL"
    done
}

# Run the daemon
main "$@"