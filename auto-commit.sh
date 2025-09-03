#!/bin/bash

# Git Air - Auto Commit Script
# Automatically commits changes with "auto commit" message

echo "🚀 Git Air - Auto Commit Starting..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not a git repository"
    exit 1
fi

# Check for changes
if git diff-index --quiet HEAD --; then
    echo "✅ No changes to commit"
    exit 0
fi

# Add all changes
echo "📝 Adding changes..."
git add .

# Check if there are staged changes
if git diff-index --quiet --cached HEAD --; then
    echo "✅ No staged changes to commit"
    exit 0
fi

# Commit with auto commit message
echo "💾 Committing changes..."
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
git commit -m "auto commit - $TIMESTAMP"

# Push to remote if configured
if git remote | grep -q origin; then
    echo "🔄 Pushing to remote..."
    git push origin HEAD
    echo "✅ Auto commit and push completed!"
else
    echo "⚠️  No remote configured, skipping push"
    echo "✅ Auto commit completed!"
fi