# addt with Daytona Provider (Experimental)

⚠️ **Experimental Feature** - The Daytona provider is under active development with limited features compared to Docker.

## Overview

The Daytona provider allows you to run Claude Code in cloud-based Daytona sandboxes instead of local Docker containers.

### Benefits
- ✅ No Docker installation required
- ✅ Cloud-based execution
- ✅ Automatic infrastructure management
- ✅ Pre-built Claude Code environment
- ✅ Sandboxes persist across sessions

### Limitations

**Not Yet Supported:**
- ❌ **Local file mounting** - Cannot access your local project files
- ❌ **Port forwarding** - Cannot run web servers and access them locally
- ❌ **GPG forwarding** - Cannot sign commits with your GPG keys
- ❌ **SSH key forwarding** - Cannot use your local SSH keys
- ❌ **Docker-in-Docker** - Cannot run Docker commands inside sandbox
- ❌ **Git config mounting** - Must configure git in sandbox manually
- ❌ **Claude config mounting** - Must use ANTHROPIC_API_KEY (no `claude login` support)
- ⚠️ **Image caching** - Limited/experimental support

**What Works:**
- ✅ Environment variables (`ADDT_ENV_VARS`)
- ✅ Persistent sandboxes
- ✅ Interactive and print modes
- ✅ All Claude Code features (model selection, continue, etc.)

## Prerequisites

1. **Daytona Account** - Sign up at [daytona.io](https://daytona.io)
2. **Daytona API Key** - Get from your Daytona dashboard
3. **Anthropic API Key** - Get from [console.anthropic.com](https://console.anthropic.com)

## Installation

Same as main addt installation - no additional setup needed. See [main README](README.md#installation).

## Quick Start

```bash
# Set up Daytona provider
export ADDT_PROVIDER=daytona
export DAYTONA_API_KEY='your-daytona-api-key'
export ANTHROPIC_API_KEY='your-anthropic-api-key'

# Run addt (creates cloud sandbox on first run)
addt "Explain how dependency injection works in Go"

# Subsequent runs reuse the sandbox (faster)
addt "Write a Python function to calculate Fibonacci"
```

## Configuration

### Required Environment Variables

| Variable | Description |
|----------|-------------|
| `ADDT_PROVIDER` | Set to `daytona` to use Daytona provider |
| `DAYTONA_API_KEY` | Your Daytona API key (get from daytona.io) |
| `ANTHROPIC_API_KEY` | Your Anthropic API key (required, no `claude login` support) |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DAYTONA_API_URL` | `https://app.daytona.io/api` | Daytona API endpoint |
| `DAYTONA_REGION` | *(org default)* | Daytona region (e.g., `us`, `eu`) |
| `ADDT_ENV_VARS` | `ANTHROPIC_API_KEY,GH_TOKEN` | Environment variables to pass to sandbox |

### Docker Variables That Don't Work

These Docker provider variables are **ignored** by Daytona:
- ❌ `ADDT_PORTS` - Port forwarding not supported
- ❌ `ADDT_GPG_FORWARD` - GPG forwarding not supported
- ❌ `ADDT_SSH_FORWARD` - SSH key forwarding not supported
- ❌ `ADDT_DOCKER_FORWARD` - Docker-in-Docker not supported

## Usage Examples

### Good Use Cases

**Code Analysis:**
```bash
export ADDT_PROVIDER=daytona
addt "Explain the difference between interfaces and abstract classes in Java"
```

**Learning & Experimentation:**
```bash
addt "Write a React component that fetches data from an API"
addt "Show me how to use goroutines in Go"
```

**Algorithm Development:**
```bash
addt "Implement a binary search tree in Python with insert and search methods"
```

### Not Suitable For

**Local Project Development:**
```bash
# ❌ Won't work - can't access local files
addt "Analyze my project structure"
addt "Fix the bug in ./src/app.js"
```

**Web Development:**
```bash
# ❌ Won't work - no port forwarding
addt "Create an Express server and let me test it"
```

**Git Operations:**
```bash
# ❌ Won't work - no SSH/GPG keys
addt "Create a signed commit and push to GitHub"
```

## How It Works

### Architecture

1. **Sandbox Creation** - On first run, creates a Daytona sandbox with Dockerfile
2. **Image Building** - Builds custom snapshot with Claude Code pre-installed
3. **SSH Connection** - Connects via SSH using Daytona API token
4. **Entrypoint Execution** - Runs entrypoint script that starts Claude Code
5. **Persistence** - Sandbox stays running for subsequent sessions

### Sandbox Lifecycle

```bash
# First run (slow - builds snapshot)
$ ADDT_PROVIDER=daytona addt "Hello"
Creating Daytona sandbox: addt-20260201-080432-88940
Building custom Daytona sandbox with Claude Code installed...
This will take a few minutes on first build...
[Building... 2-3 minutes]
Sandbox is ready!
Creating Daytona API client...
Connecting via SSH...

# Second run (fast - reuses sandbox)
$ ADDT_PROVIDER=daytona addt "Continue"
Using existing Daytona sandbox: addt-20260201-080432-88940
Creating Daytona API client...
Connecting via SSH...
[Instant connection]
```

### Naming Convention

**Ephemeral Mode (default):**
- Format: `addt-YYYYMMDD-HHMMSS-PID`
- Example: `addt-20260201-080432-88940`
- Removed after exit

**Persistent Mode:**
```bash
export ADDT_PERSISTENT=true
# Format: addt-workspace (or custom name)
```

## Troubleshooting

### Authentication Errors

**Problem:** "Forbidden: Forbidden resource"
```bash
Error: failed to create snapshot: exit status 1
FATA[0000] Forbidden: Forbidden resource
```

**Solution:** Check your Daytona API key and permissions:
```bash
# Verify API key is set
echo $DAYTONA_API_KEY

# Check if you can list sandboxes
daytona sandbox list

# Verify organization permissions
daytona profile list
```

### Sandbox Not Found

**Problem:** Can't connect to existing sandbox

**Solution:** The sandbox may have been stopped or deleted:
```bash
# List your sandboxes
daytona sandbox list

# Create a new sandbox (removes old naming)
unset ADDT_PERSISTENT
ADDT_PROVIDER=daytona addt
```

### Slow First Build

**Problem:** First run takes 2-3 minutes

**Explanation:** This is expected - Daytona builds a custom snapshot with Claude Code.

**Subsequent runs are fast** (~10-20 seconds) because they reuse the snapshot.

### Can't Access Local Files

**Problem:** Claude can't see my project files

**Solution:** This is a known limitation. Options:
1. Use Docker provider for local projects
2. Manually upload files to sandbox
3. Work with code snippets instead of files

### Connection Timeout

**Problem:** SSH connection times out

**Solution:**
```bash
# Check your internet connection
ping app.daytona.io

# Verify sandbox is running
daytona sandbox info <sandbox-name>

# Restart sandbox if needed
daytona sandbox stop <sandbox-name>
daytona sandbox start <sandbox-name>
```

## Comparison: Daytona vs Docker

| Feature | Docker | Daytona |
|---------|--------|---------|
| **Local Files** | ✅ Full access | ❌ Not available |
| **Port Forwarding** | ✅ Automatic | ❌ Not supported |
| **SSH Forwarding** | ✅ Agent & keys | ⚠️ Built-in only |
| **GPG Forwarding** | ✅ Sign commits | ❌ Not supported |
| **Docker-in-Docker** | ✅ Isolated/host | ❌ Not supported |
| **Git Config** | ✅ Auto-mounted | ❌ Manual setup |
| **Claude Config** | ✅ Auto-mounted | ❌ API key only |
| **Persistent State** | ✅ Per directory | ✅ Per sandbox |
| **Offline Mode** | ✅ Works offline | ❌ Requires internet |
| **Setup Required** | Docker install | Daytona account |

**Use Docker for:**
- ✅ Local project development
- ✅ Web application development
- ✅ Full-featured workflow
- ✅ Offline work
- ✅ Production use

**Use Daytona for:**
- ✅ Quick experimentation
- ✅ Learning and tutorials
- ✅ Code snippets and algorithms
- ✅ When Docker isn't available
- ⚠️ Non-production use only

## Future Development

Planned features for Daytona provider:
- [ ] Local file mounting via Daytona volumes
- [ ] Port forwarding support
- [ ] SSH key forwarding
- [ ] Git config synchronization
- [ ] Image caching improvements
- [ ] Better error messages

## Contributing

The Daytona provider is experimental. Contributions welcome! See [README-development.md](README-development.md) for details.

## Support

- **Docker Provider Issues:** Use the main [issue tracker](https://github.com/jedi4ever/addt/issues)
- **Daytona Provider Issues:** Label issues with `provider:daytona`
- **Daytona Platform Issues:** Contact [Daytona support](https://daytona.io)

## Links

- [Main README](README.md)
- [Development Guide](README-development.md)
- [Daytona Documentation](https://www.daytona.io/docs)
- [Daytona CLI Reference](https://www.daytona.io/docs/en/tools/cli/)
