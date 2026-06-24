# wtool - Universal Git Worktree Automation Tool

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

**wtool** is a zero-config automation tool that solves git worktree compatibility issues with AI coding assistants (Codex, Claude, Cursor, Qoder, Trae).

## Problem

When using git worktrees with AI coding tools, you encounter:

- **Submodules** not auto-initialized in worktrees
- **Codegraph/Graphify** storing absolute paths that break across worktrees
- **pnpm** workspace root detection failing in worktrees
- **AI tools** each handling worktrees differently with no unified approach

## Solution

wtool auto-detects and fixes all these issues with a single command.

## Features

- **Zero-config** - Works out of the box with sensible defaults
- **Auto-detection** - Detects submodules, pnpm workspaces, codegraph, and AI tool configs
- **Smart submodule init** - Prefers local clone from primary worktree to avoid re-downloading
- **Path repair** - Converts absolute paths to relative in graph files and configs
- **Dependency linking** - Reuses `node_modules` from primary worktree when safe
- **Port allocation** - Auto-assigns ports based on worktree index to avoid conflicts
- **AI tool integration** - Generates configs for Claude, Cursor, Codex, Qoder, Trae
- **Git hooks** - Auto-runs on checkout/merge for seamless workflow

## Installation

### One-line installer (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/Danceiny/wtool/main/scripts/install.sh | bash
```

### Go install

```bash
go install github.com/Danceiny/wtool@latest
```

### Build from source

```bash
git clone https://github.com/Danceiny/wtool.git
cd wtool
make build
make install
```

## Quick Start

```bash
# Check your environment
wtool doctor

# Initialize current worktree (auto-detect and fix everything)
wtool init

# List all worktrees
wtool list

# Install git hooks for auto-initialization
wtool setup hook --global

# Dry-run to preview changes
wtool init --dry-run
```

## Commands

| Command | Description |
|---------|-------------|
| `wtool init [path]` | Initialize worktree with auto-detection |
| `wtool sync` | Sync worktree state (submodules, paths, deps) |
| `wtool doctor` | Diagnose environment and detect issues |
| `wtool list` | List all worktrees |
| `wtool env` | Output worktree environment variables |
| `wtool setup hook` | Install git hooks |
| `wtool setup config` | Show AI tool configuration examples |

## AI Tool Integration

### Claude

Add to `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      "wtool init --quiet"
    ]
  }
}
```

### Cursor

Add to `.cursor/worktrees.json`:

```json
{
  "setup-worktree": ["wtool init"]
}
```

### Codex

Add to `AGENTS.md`:

```markdown
## Worktree Setup
Run `wtool init` before starting work in a new worktree.
```

### Qoder

Configure in `.qoder/settings.local.json` to run `wtool init` on session start.

## Configuration

wtool uses a layered configuration system (in order of priority):

1. Command-line flags
2. `.wtool.yaml` in current directory
3. `~/.config/wtool/config.yaml`
4. Environment variables (`WTOOL_*`)
5. Auto-detected defaults

### Example `.wtool.yaml`

```yaml
submodules:
  enabled: true
  strategy: "local-clone-first"  # or "remote"
  
ports:
  enabled: true
  services:
    - name: "api"
      base: 8080
      env_key: "API_PORT"
    - name: "frontend"
      base: 3000
      env_key: "FE_PORT"

path_fixes:
  enabled: true
  
ai_tools:
  claude:
    enabled: true
  cursor:
    enabled: true
```

## How It Works

### Submodule Initialization

1. Detects `.gitmodules` and parses submodule paths
2. Checks if each submodule is already initialized
3. Tries to clone from primary worktree's local checkout (fast, no network)
4. Falls back to `git submodule update --init` if local clone fails

### Path Repair

1. Scans for `.codegraph/codegraph.db`, `graph.json`, and config files
2. Detects absolute paths stored in these files
3. Rewrites them to relative paths based on worktree root

### Dependency Linking

1. Checks if `package.json` and lock files match between worktrees
2. If identical, creates a symlink to primary worktree's `node_modules`
3. If different or no source available, runs `pnpm install` or equivalent

### Port Allocation

1. Gets worktree index from `git worktree list`
2. Adds index to base port numbers
3. Generates `.worktree.env` with assigned ports

## License

MIT License - see [LICENSE](LICENSE) file.
