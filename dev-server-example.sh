#!/bin/bash

# Git Air - Dev Server Setup Example
# This shows typical usage on a development server

echo "🏠 Git Air Dev Server Setup Example"
echo "====================================="

echo ""
echo "📂 Your development server structure might look like:"
echo "~/home/user/"
echo "├── project1/          (Git repo with origin + backup remotes)"
echo "├── project2/          (Git repo with multiple remotes)"
echo "├── clients/"
echo "│   ├── client-a/      (Git repo)"
echo "│   └── client-b/      (Git repo)"
echo "├── experiments/"
echo "│   ├── ai-tool/       (Git repo)"
echo "│   └── web-scraper/   (Git repo)"
echo "└── git-air            (The git-air binary)"

echo ""
echo "🚀 To start Git Air and manage ALL projects:"
echo "cd ~/home/user"
echo "./git-air"

echo ""
echo "✨ Git Air will then:"
echo "• Discover all 6 Git repositories automatically"
echo "• Monitor for changes in real-time"
echo "• Auto-commit any changes with timestamp"
echo "• Push to ALL configured remotes for each repo"
echo "• Pull updates from remotes every minute"
echo "• Handle new repositories added later"

echo ""
echo "🔧 Configuration options:"
echo "• Create git-air.yaml for custom settings"
echo "• Use -log debug for detailed output"
echo "• Use -scan \"path1,path2\" for specific paths"

echo ""
echo "🛡️ Safety features:"
echo "• Excludes .git, node_modules, temp files"
echo "• Handles merge conflicts gracefully"
echo "• Detailed logging of all operations"
echo "• Graceful shutdown with Ctrl+C"

echo ""
echo "💡 Perfect for:"
echo "• Development servers with multiple projects"
echo "• Team environments needing auto-sync"
echo "• Backup automation to multiple remotes"
echo "• Never losing work due to automatic commits"