# addt - Development Guide

Technical documentation for developers and contributors.

## Table of Contents

- [Architecture](#architecture)
- [Build System](#build-system)
- [Docker Image Structure](#docker-image-structure)
- [Volume Mounts](#volume-mounts)
- [Image Metadata & Labels](#image-metadata--labels)
- [File Structure](#file-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Advanced Usage](#advanced-usage)

## Architecture

### Base Image

addt uses `node:${NODE_VERSION}-slim` as the base image:
- **Debian-based** for easy package installation
- **Slim variant** for smaller image size (~500MB vs 1GB+)
- **Configurable Node version** via `ADDT_NODE_VERSION`

### Non-Root User Setup

The container runs as a non-root user matching your local user:
- **UID/GID matching**: Container user has same UID/GID as host user
- **Benefit**: Files created in container have correct ownership on host
- **Implementation**: Build args pass `USER_ID` and `GROUP_ID` from `id -u` and `id -g`

### Installed Tools

Tools pre-installed in the image:
- **Claude Code** - Latest or pinned version from npm
- **Git** - Version control
- **GitHub CLI (gh)** - GitHub operations
- **Ripgrep (rg)** - Fast code search
- **Docker CLI** - For Docker-in-Docker support
- **curl** - HTTP requests
- **sudo** - For privileged operations (DinD)

### Volume Mounting

The wrapper script (`addt.sh`) automatically mounts:

1. **Current directory** → `/workspace`
   - Your project files
   - Read/write access

2. **`~/.gitconfig`** → `/home/<user>/.gitconfig` (read-only)
   - Git identity (name, email)
   - Git aliases and configuration
   - Prevents accidental modification

3. **`~/.claude`** → `/home/<user>/.claude`
   - Session persistence
   - Conversation history
   - Authentication credentials

4. **`~/.claude.json`** → `/home/<user>/.claude.json`
   - Claude configuration
   - Preferences

5. **`~/.gnupg`** → `/home/<user>/.gnupg` (opt-in)
   - GPG keys for commit signing
   - Only when `ADDT_GPG_FORWARD=true`

6. **`~/.ssh`** or SSH agent (opt-in)
   - SSH keys for git operations
   - Agent forwarding or key mounting
   - Only when `ADDT_SSH_FORWARD` is set

### Authentication & Identity

**Claude Code Authentication:**
- **Option 1**: Mount `~/.claude` directory (automatic if exists)
- **Option 2**: Pass `ANTHROPIC_API_KEY` environment variable

**Git Identity:**
- Automatically inherited from host `~/.gitconfig`
- No manual configuration needed
- Commits show your name/email

**GitHub CLI:**
- Requires `GH_TOKEN` environment variable
- Or use `ADDT_GITHUB_DETECT=true` to auto-detect from `gh` CLI

### Image Metadata & Labels

Images include OCI-compliant labels:

```json
{
  "org.opencontainers.image.title": "addt",
  "org.opencontainers.image.description": "Claude Code with Git, GitHub CLI, and Ripgrep",
  "tools.claude.version": "2.1.27",
  "tools.gh.version": "2.86.0",
  "tools.ripgrep.version": "13.0.0",
  "tools.git.version": "2.39.5",
  "tools.node.version": "20.20.0"
}
```

**Query images by labels:**
```bash
# Filter images by tool
docker images --filter "label=tools.claude.version=2.1.27"

# Get specific label value
docker inspect addt:claude-2.1.27 --format '{{index .Config.Labels "tools.node.version"}}'

# View all labels
docker inspect addt:latest --format '{{json .Config.Labels}}' | jq
```

## Build System

### Standalone Build

The `make standalone` command creates a single self-contained script:

```bash
make standalone
# Creates: dist/addt-standalone.sh
```

**How it works:**
1. `build.sh` reads `addt.sh`, `Dockerfile`, and `docker-entrypoint.sh`
2. Embeds Dockerfile and entrypoint as heredocs at `INJECT:*` markers
3. Adds cleanup for temporary files
4. Outputs single executable script

**Build process:**
```bash
# build.sh uses awk to inject content
awk '
/INJECT:Dockerfile-start/ {
    print "    DOCKERFILE_PATH=\"$SCRIPT_DIR/.addt-Dockerfile.tmp\""
    print "    cat > \"$DOCKERFILE_PATH\" <<'\''DOCKERFILE_EOF'\''"
    # Insert full Dockerfile content
    print "DOCKERFILE_EOF"
    # Skip until end marker
    while (getline > 0 && !/INJECT:Dockerfile-end/) { }
    next
}
{ print }
' addt.sh > dist/addt-standalone.sh
```

### Version Management

**Automatic version detection:**
```bash
# Query npm registry via HTTP (no npm CLI needed)
NPM_LATEST=$(curl -s https://registry.npmjs.org/@anthropic-ai/claude-code | grep -o '"stable":"[^"]*"' | cut -d'"' -f4)
```

**Version validation:**
```bash
# Check if version exists in registry
NPM_DATA=$(curl -s https://registry.npmjs.org/@anthropic-ai/claude-code)
echo "$NPM_DATA" | grep -q "\"$VERSION\":"
```

**Image reuse logic:**
1. Check if image with version label exists
2. If yes, use existing image (skip rebuild)
3. If no, build new image with version tag

### Port Mapping System

**Automatic port allocation:**
```bash
# Check port availability using bash built-in
is_port_available() {
    local port=$1
    (bash -c "exec 3<>/dev/tcp/localhost/$port" 2>/dev/null && exec 3>&-) && return 1 || return 0
}

# Find next available port
find_available_port() {
    local start_port=$1
    local port=$start_port
    while ! is_port_available "$port"; do
        port=$((port + 1))
    done
    echo "$port"
}
```

**Port mapping passed to Claude:**
```bash
# In docker-entrypoint.sh
SYSTEM_PROMPT="Port mappings (container→host):
- Container port 3000 → Host port 30000
When telling the user URLs, use host ports."

exec claude --append-system-prompt "$SYSTEM_PROMPT" "$@"
```

## Docker Image Structure

### Multi-stage Build Context

The Dockerfile uses build arguments for flexibility:

```dockerfile
ARG NODE_VERSION=20
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USERNAME=claude
ARG CLAUDE_VERSION=latest
```

### User Management

```dockerfile
# Create user with matching UID/GID
RUN (groupadd -g ${GROUP_ID} ${USERNAME} 2>/dev/null || true) \
    && useradd -m -u ${USER_ID} -g ${GROUP_ID} -s /bin/bash ${USERNAME} \
    && echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
```

### Entrypoint Wrapper

`docker-entrypoint.sh` handles:
1. Docker-in-Docker daemon startup (if enabled)
2. Port mapping system prompt injection
3. Forwarding to Claude Code CLI

## File Structure

```
addt/
├── addt.sh              # Main wrapper script (source)
├── Dockerfile              # Container definition
├── docker-entrypoint.sh    # Container entrypoint
├── build.sh                # Standalone builder
├── Makefile                # Build automation
├── .dockerignore           # Build context exclusions
├── .gitignore              # Git exclusions
├── README.md               # User documentation
├── README-development.md   # This file
└── dist/
    └── addt-standalone.sh  # Single-file distribution
```

## Development Workflow

### Local Development

```bash
# Make changes to addt.sh, Dockerfile, or docker-entrypoint.sh
vim addt.sh

# Test changes
./addt.sh --version

# Rebuild standalone
make standalone

# Test standalone
./dist/addt-standalone.sh --version
```

### Force Rebuild

```bash
# Remove image to force rebuild
docker rmi addt:claude-2.1.27

# Or remove all addt images
docker images | grep addt | awk '{print $3}' | xargs docker rmi

# Next run will rebuild
./addt.sh --version
```

### Debug Mode

```bash
# Enable logging
export ADDT_LOG=true
export ADDT_LOG_FILE="/tmp/addt-debug.log"
./addt.sh

# View logs
tail -f /tmp/addt-debug.log
```

### Shell Access

```bash
# Open bash shell in container
./addt.sh shell

# Run specific command
./addt.sh shell -c "env | grep DCLAUDE"

# Check mounted directories
./addt.sh shell -c "ls -la ~ && ls -la /workspace"
```

## Testing

### Test Suite

```bash
# Test basic functionality
./addt.sh --version

# Test port mapping
ADDT_PORTS="3000,8080" ./addt.sh --version

# Test SSH forwarding
ADDT_SSH_FORWARD=agent ./addt.sh shell -c "ssh-add -l"

# Test Docker forwarding
ADDT_DOCKER_FORWARD=isolated ./addt.sh shell -c "docker ps"

# Test GPG forwarding
ADDT_GPG_FORWARD=true ./addt.sh shell -c "gpg --list-keys"
```

### Version Tests

```bash
# Test invalid version
ADDT_CLAUDE_VERSION=2.1.24 ./addt.sh --version
# Should error: version does not exist

# Test specific version
ADDT_CLAUDE_VERSION=2.1.29 ./addt.sh --version
# Should build with 2.1.29

# Test version reuse
ADDT_CLAUDE_VERSION=2.1.29 ./addt.sh --version
# Should reuse existing image (fast)
```

### Standalone Tests

```bash
# Build standalone
make standalone

# Test in clean directory
cd /tmp
/path/to/addt/dist/addt-standalone.sh --version

# Test embedded Dockerfile
./dist/addt-standalone.sh shell -c "ls -la .addt-*.tmp"
```

## Advanced Usage

### Custom Dockerfile

To modify the image, edit `Dockerfile`:

```dockerfile
# Add additional tools
RUN apt-get update && apt-get install -y \
    vim \
    htop \
    && apt-get clean

# Install additional npm packages
RUN npm install -g typescript prettier eslint
```

Then rebuild:
```bash
docker rmi addt:latest
./addt.sh  # Auto-rebuilds
```

### Alpine-based Image

For a smaller image, modify `Dockerfile` base:

```dockerfile
FROM node:${NODE_VERSION}-alpine
RUN apk add --no-cache \
    git \
    curl \
    bash \
    # ... other packages
```

**Note:** Alpine uses `apk` instead of `apt-get` and may have compatibility issues with some npm packages.

### Multi-Architecture Builds

Build for multiple platforms:

```bash
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --build-arg CLAUDE_VERSION=2.1.27 \
    -t addt:latest \
    .
```

### Environment Variable Injection

Pass custom environment variables:

```bash
export ADDT_ENV_VARS="ANTHROPIC_API_KEY,AWS_ACCESS_KEY_ID,AWS_SECRET_ACCESS_KEY"
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
./addt.sh
```

Inside container, Claude can access:
```bash
echo $AWS_ACCESS_KEY_ID  # Available
```

## Contributing

### Code Style

- Use 4 spaces for indentation in shell scripts
- Add comments for non-obvious logic
- Keep functions focused and single-purpose
- Use descriptive variable names

### Commit Guidelines

- Use conventional commits format
- Include Claude Code attribution:
  ```
  feat: add port mapping support

  Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
  ```

### Pull Request Process

1. Test changes locally
2. Update documentation (README.md and README-development.md)
3. Rebuild standalone: `make standalone`
4. Create PR with clear description
5. Include test commands in PR description

### Testing Checklist

Before submitting PR:
- [ ] `./addt.sh --version` works
- [ ] `make standalone` succeeds
- [ ] `./dist/addt-standalone.sh --version` works
- [ ] Port mapping works (if modified)
- [ ] Volume mounts work (test with `./addt.sh shell`)
- [ ] Documentation updated

## Debugging

### Common Issues

**Build fails:**
```bash
# Check Docker daemon
docker info

# Check disk space
df -h

# Clean Docker cache
docker system prune -a
```

**Volume mount issues:**
```bash
# Check permissions
ls -la ~/.claude
ls -la ~/.gitconfig

# Test mount
./addt.sh shell -c "ls -la ~"
```

**Port mapping issues:**
```bash
# Check port availability
bash -c "exec 3<>/dev/tcp/localhost/30000" && echo "Port busy" || echo "Port free"

# Test with specific ports
ADDT_PORTS="3000" ADDT_PORT_RANGE_START=40000 ./addt.sh --version
```

### Verbose Mode

```bash
# Enable bash debugging
bash -x ./addt.sh --version 2>&1 | tee debug.log

# Or modify script temporarily
set -x  # Add to top of addt.sh
```

### Container Inspection

```bash
# List running containers
docker ps -a | grep addt

# Inspect container
docker inspect addt-YYYYMMDD-HHMMSS-PID

# View logs
docker logs addt-YYYYMMDD-HHMMSS-PID
```

## License

MIT License - See LICENSE file for details.
