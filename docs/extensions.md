# addtExtensions

Extensions allow you to add tools and AI agents to your addtcontainer image. The base image provides infrastructure (Node.js, Go, Python/UV, Git, GitHub CLI), and extensions add the actual tools.

## Available Extensions

### AI Coding Agents

| Extension | Description | Entrypoint | Provider |
|-----------|-------------|------------|----------|
| `claude` | Claude Code - AI coding assistant | `claude` | Anthropic |
| `codex` | OpenAI Codex CLI - AI coding assistant | `codex` | OpenAI |
| `cursor` | Cursor CLI Agent - AI-powered code editor agent | `cursor` / `agent` | Cursor |
| `amp` | Amp - AI coding agent | `amp` | Sourcegraph |
| `gemini` | Gemini CLI - AI coding agent | `gemini` | Google |
| `copilot` | GitHub Copilot CLI - AI coding assistant | `copilot` | GitHub |

### Utility Extensions

| Extension | Description | Entrypoint | Dependencies |
|-----------|-------------|------------|--------------|
| `beads` | Git-backed issue tracker for AI agents | `bd` | - |
| `gastown` | Multi-agent orchestration for Claude Code | `gt` | claude, beads |
| `tessl` | Agent enablement platform - package manager for AI agent skills | `tessl` | - |

**Note:** The `claude` extension is installed by default. When you build with other extensions like `gastown`, their dependencies (including `claude`) are automatically installed.

## Using Extensions

### Building with Extensions

Use the `containers build` command with `--build-arg` to include extensions:

```bash
# Default build (installs claude extension)
addtcontainers build

# Build with gastown (automatically includes claude and beads dependencies)
addtcontainers build --build-arg ADDT_EXTENSIONS=gastown

# Build with multiple extensions
addtcontainers build --build-arg ADDT_EXTENSIONS=claude,tessl

# Build minimal image with only tessl (no claude)
addtcontainers build --build-arg ADDT_EXTENSIONS=tessl

# Via environment variable
ADDT_EXTENSIONS=gastown addtcontainers build
```

### Image Naming Convention

Docker images are automatically named based on the installed extensions and their versions:

```bash
# Single extension
addt:claude-2.1.17

# Multiple extensions (sorted alphabetically)
addt:claude-2.1.17_codex-latest

# Different combination = different image
addt:gemini-latest_tessl-latest
```

This ensures that different extension combinations always get their own isolated images.

### Extension Dependencies

Extensions can depend on other extensions. Dependencies are automatically resolved and installed in the correct order.

For example, `gastown` depends on `beads`, so running:

```bash
addtbuild --build-arg ADDT_EXTENSIONS=gastown
```

Will automatically install both `beads` and `gastown`.

### Checking Installed Extensions

After building, you can verify installed extensions:

```bash
# Check extension metadata
addtshell -c "cat ~/.addt/extensions.json"

# Check specific tools
addtshell -c "which gt bd tessl"
```

### Symlink-Based Extension Selection

You can create symlinks to the `addt` binary with names matching your extensions. When invoked via a symlink, addtautomatically uses that extension:

```bash
# Create symlinks
ln -s addtcodex
ln -s addtgemini
ln -s addtclaude-flow

# Now these are equivalent:
./codex "help me with this code"           # Uses codex extension
ADDT_EXTENSIONS=codex addt"..."     # Same result

./gemini "explain this function"           # Uses gemini extension
ADDT_EXTENSIONS=gemini addt"..."    # Same result
```

**How it works:**
- Detects the binary name from how it was invoked
- If not "addt", sets `ADDT_EXTENSIONS` and `ADDT_COMMAND` to match the binary name
- Environment variables can still override this behavior

This is useful for:
- Creating dedicated commands for different AI agents
- Simplifying workflows when you frequently use a specific agent
- Installing multiple "binaries" from a single addtinstallation

### Per-Extension Configuration

Each extension can be configured individually via environment variables:

```bash
# Set version for a specific extension
ADDT_CLAUDE_VERSION=2.0.0 addtcontainers build
ADDT_CODEX_VERSION=0.1.0 addtcontainers build

# Disable config directory mounting for an extension
ADDT_CLAUDE_MOUNT_CONFIG=false addt

# Multiple extensions with specific versions
ADDT_EXTENSIONS=claude,codex \
  ADDT_CLAUDE_VERSION=2.1.0 \
  ADDT_CODEX_VERSION=latest \
  addtcontainers build
```

| Variable Pattern | Description |
|-----------------|-------------|
| `ADDT_<EXT>_VERSION` | Version to install (e.g., `2.1.0`, `latest`, `stable`) |
| `ADDT_<EXT>_MOUNT_CONFIG` | Mount extension config dirs (`true`/`false`) |

### Automatic Environment Variable Forwarding

Extensions can declare which environment variables they need in their `config.yaml`. When running addt, these variables are automatically forwarded from your host to the container - no need to specify them manually.

**Example extension configs:**

```yaml
# claude extension
env_vars:
  - ANTHROPIC_API_KEY

# codex extension
env_vars:
  - OPENAI_API_KEY

# gemini extension
env_vars:
  - GEMINI_API_KEY
  - GOOGLE_API_KEY
```

**How it works:**

1. When you build an image, each extension's `env_vars` are collected into `~/.addt/extensions.json`
2. At runtime, addtreads this metadata and automatically forwards listed variables from host to container
3. Variables are only forwarded if they're set on the host (empty values are skipped)

**Benefits:**

- No need to remember which API keys each tool needs
- Just set the variable on your host once, it's automatically available in containers
- Different extensions in the same image can have different env vars
- Users can still add additional variables via `ADDT_FORWARD_ENV`

**Example:**

```bash
# Just set your API keys on the host
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."

# Build with both extensions
addtcontainers build --build-arg ADDT_EXTENSIONS=claude,codex

# Run - both API keys are automatically forwarded
addt"help me with this code"        # Uses ANTHROPIC_API_KEY
ADDT_COMMAND=codex addt"..."     # Uses OPENAI_API_KEY
```

## Creating Extensions

Extensions are stored in `src/assets/docker/extensions/` as directories containing:

```
extensions/
└── myextension/
    ├── config.yaml    # Extension metadata (required)
    ├── install.sh     # Installation script (optional, runs at build time)
    └── setup.sh       # Setup script (optional, runs at container startup)
```

**Note:** Only `config.yaml` is required. Extensions can be metadata-only (no install.sh or setup.sh) if they just need to define mounts or dependencies.

### config.yaml

Defines extension metadata:

```yaml
name: myextension
description: Short description of what the extension does
entrypoint: mycommand
dependencies:
  - beads           # Other extensions this depends on
env_vars:
  - MY_API_KEY      # Environment variables to forward from host
  - MY_SECRET_TOKEN
mounts:
  - source: ~/.myextension
    target: /home/claude/.myextension
  - source: ~/.config/myextension
    target: /home/claude/.config/myextension
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Extension identifier (should match directory name) |
| `description` | Yes | Brief description |
| `entrypoint` | Yes | Main command provided by extension |
| `dependencies` | No | List of other extensions required |
| `env_vars` | No | Environment variables to automatically forward from host |
| `mounts` | No | Directories to mount from host to container |

### Extension Files

| File | Required | When it runs | Description |
|------|----------|--------------|-------------|
| `config.yaml` | Yes | Build time | Extension metadata and configuration |
| `install.sh` | No | Build time | Installs packages and tools into the image |
| `setup.sh` | No | Runtime | Runs at container startup for initialization |

### Mounts

Extensions can specify directories to be mounted from the host into the container at runtime. This is useful for:

- Persisting extension configuration across container restarts
- Sharing data between host and container
- Caching extension data

Each mount entry requires:
- `source`: Path on the host (supports `~` for home directory)
- `target`: Path inside the container

The host directories are automatically created if they don't exist.

### install.sh

The installation script runs during Docker image build. It has access to:

- **apt** (via `sudo`) - for system packages
- **npm** (via `sudo`) - for Node.js packages
- **go** - for Go packages (installed to `~/go/bin`)
- **pip/uv** - for Python packages

Example install script:

```bash
#!/bin/bash
set -e

echo "Extension [myextension]: Installing..."

# System packages (requires sudo)
sudo apt-get update && sudo apt-get install -y --no-install-recommends \
    some-package

# Node.js packages (requires sudo for global)
sudo npm install -g @some/package

# Go packages (no sudo needed, installs to ~/go/bin)
/usr/local/go/bin/go install github.com/user/repo/cmd/tool@latest

# Python packages
uv pip install some-package

echo "Extension [myextension]: Done."
```

### setup.sh (Optional)

The setup script runs at container startup (runtime), not during image build. Use it for:

- Initializing runtime state
- Displaying welcome messages
- Checking for required environment variables
- Setting up runtime configuration

Example setup script:

```bash
#!/bin/bash
echo "Setup [myextension]: Initializing environment"

# Check for required API key
if [ -z "$MY_API_KEY" ]; then
    echo "Warning: MY_API_KEY not set"
fi
```

Setup scripts run once per container session. In persistent mode, they only run on the first start (a marker file prevents re-running).

### Testing Your Extension

1. Create the extension directory and files
2. Build addt: `make build`
3. Build image with extension: `./dist/addtcontainers build --build-arg ADDT_EXTENSIONS=myextension`
4. Verify: `./dist/addtshell -c "which mycommand"`

## Extension Metadata

When extensions are installed, metadata is written to `~/.addt/extensions.json`:

```json
{
  "extensions": {
    "claude": {
      "name": "claude",
      "description": "Claude Code - AI coding assistant by Anthropic",
      "entrypoint": "claude",
      "mounts": [
        {"source": "~/.claude", "target": "/home/claude/.claude"},
        {"source": "~/.claude.json", "target": "/home/claude/.claude.json"}
      ],
      "flags": [
        {"flag": "--yolo", "description": "Bypass permission checks"}
      ],
      "env_vars": ["ANTHROPIC_API_KEY"]
    },
    "gastown": {
      "name": "gastown",
      "description": "Multi-agent orchestration for Claude Code",
      "entrypoint": "gt",
      "mounts": [
        {"source": "~/.gastown", "target": "/home/claude/.gastown"}
      ],
      "env_vars": ["ANTHROPIC_API_KEY"]
    }
  }
}
```

This metadata is used at runtime to:
- Mount extension directories from the host
- Discover available extensions and their entrypoints
- Automatically forward required environment variables

## Examples

### AI Coding Agents

You can build images with different AI coding agents and switch between them:

```bash
# Build with multiple AI agents
addtcontainers build --build-arg ADDT_EXTENSIONS=claude,codex,gemini,copilot

# Run Claude (default)
addt

# Run OpenAI Codex
ADDT_COMMAND=codex addt

# Run Google Gemini
ADDT_COMMAND=gemini addt

# Run GitHub Copilot
ADDT_COMMAND=copilot addt

# Run Sourcegraph Amp
ADDT_COMMAND=amp addt

# Run Cursor Agent
ADDT_COMMAND=cursor addt
```

**Using symlinks for dedicated agent commands:**

```bash
# Create symlinks for each agent
cd /usr/local/bin  # or wherever addtis installed
ln -s addtcodex
ln -s addtgemini
ln -s addtcopilot

# Build images for each (first run will auto-build)
codex containers build
gemini containers build

# Now use them directly
codex "refactor this function"
gemini "explain this code"
```

Each symlink automatically builds and uses its own isolated image (`addt:codex-latest`, `addt:gemini-latest`, etc.).

### Cursor Extension

Cursor CLI provides an AI-powered code editor agent:

```bash
# Build with cursor only
addtcontainers build --build-arg ADDT_EXTENSIONS=cursor

# Run cursor agent
ADDT_COMMAND=cursor addt
# or
ADDT_COMMAND=agent addt

# Or use symlink
ln -s addtcursor
./cursor "help me with this code"
```

### Gastown Extension

Gastown provides multi-agent orchestration for Claude Code:

```bash
# Build with gastown
addtcontainers build --build-arg ADDT_EXTENSIONS=gastown

# Run gastown instead of claude
ADDT_COMMAND=gt addt

# Or use shell mode
addtshell
gt --help
```

### Tessl Extension

Tessl is an agent enablement platform with a skills package manager:

```bash
# Build with tessl
addtcontainers build --build-arg ADDT_EXTENSIONS=tessl

# Use tessl
addtshell
tessl init           # Authenticate
tessl skill search   # Find skills
tessl mcp            # Start MCP server
```

## Troubleshooting

### Permission Errors

If you see permission errors during installation:
- Use `sudo` for `apt-get` and global `npm install`
- Go packages don't need sudo (install to user's `~/go/bin`)

### Extension Not Found

If an extension is not recognized:
- Ensure the directory name matches the extension name in `config.yaml`
- Check that `config.yaml` exists (install.sh and setup.sh are optional)
- Rebuild addtwith `make build` to embed the new extension
