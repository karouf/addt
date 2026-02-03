# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.7.1] - 2026-02-01

### Fixed
- Fixed version mismatch false positive when using dist-tags (stable, latest, next)
- DetermineImageName() now updates config with resolved version before checking for existing images
- Eliminates unnecessary rebuilds when using 'stable' tag
- Existing images are now properly reused without rebuild

### Technical
- Moved ClaudeVersion resolution before existing image check in DetermineImageName()
- Added 'next' dist-tag to version resolution logic

## [1.7.0] - 2026-02-01

### Added
- Homebrew installation instructions as primary installation method (Option 1)
- Support for "latest" tag in Go version (queries https://go.dev/VERSION?m=text)
- Support for npm dist-tags (stable, latest, next) for Claude Code
- UV self-update capability via official install script

### Changed
- **Claude Code default changed from "latest" to "stable"** (2.1.29 → 2.1.17)
  - More stable for production use
  - Users can still use `latest` or `next` if needed
- **UV installation now uses official install script** instead of manual binary download
  - Enables `uv self update` inside containers
  - Installs to ~/.local/bin as claude user
- **UV default version changed to "latest"** (was 0.5.11, now gets 0.9.28+)
- **Go default version changed to "latest"** (was 1.23.5, now gets 1.25.6+)
  - Automatically queries go.dev for newest stable version
- Updated all documentation to reflect new defaults

### Installation
- Homebrew tap available: `brew tap jedi4ever/tap && brew install addt`
- Added instructions for tap setup, upgrade, and versioned installs

### Technical
- Added `getNpmVersionByTag()` to fetch versions for any npm dist-tag
- Modified `DetermineImageName()` to handle dist-tags (stable, latest, next)
- Dockerfile now queries go.dev/VERSION for latest Go version when GO_VERSION=latest
- UV install script supports version-in-URL format (https://astral.sh/uv/{version}/install.sh)

### Current Defaults
- ADDT_CLAUDE_VERSION=stable (2.1.17)
- ADDT_UV_VERSION=latest (0.9.28+)
- ADDT_GO_VERSION=latest (1.25.6+)
- ADDT_NODE_VERSION=20

### Tested
- Go 1.25.6 installs correctly with "latest"
- UV 0.9.28 installs correctly with "latest"
- `uv self update` works inside containers
- Claude Code stable tag resolves to 2.1.17
- All three dist-tags work (stable, latest, next)

## [1.6.0] - 2026-02-01

### Added
- Network firewall with whitelist-based domain filtering using iptables and ipset
- Firewall management commands: `addt firewall list|add|remove|reset`
- Three firewall modes: strict (block non-whitelisted), permissive (log only), off
- Default whitelist includes: Anthropic API, GitHub, npm, PyPI, Go modules, Docker Hub, CDNs
- Configuration via ADDT_FIREWALL and ADDT_FIREWALL_MODE environment variables
- Firewall config file at ~/.addt/firewall/allowed-domains.txt
- Automatic NET_ADMIN capability when firewall is enabled
- Firewall status display in container status line
- Firewall initialization in both run and shell modes
- Test script for firewall functionality verification

### Changed
- Docker provider now mounts firewall config directory when firewall is enabled
- Shell mode wrapper now initializes firewall before opening bash
- Updated help text and README with firewall documentation
- Added Credits section in README acknowledging claude-clamp inspiration

### Technical
- Added init-firewall.sh script for firewall initialization
- Added firewall.go for domain management commands
- Firewall resolves domain names to IPs using dig/host commands
- ipset stores up to 65536 whitelisted IPs with 4096 hashsize
- iptables rules: ACCEPT allowed_ips, LOG blocked traffic, DROP everything else

### Tested
- Firewall blocks non-whitelisted domains (google.com, example.com)
- Firewall allows whitelisted domains (api.anthropic.com)
- Domain management commands work correctly
- Firewall initialization works in both run and shell modes
- Persistent and ephemeral container modes both supported

## [1.5.0] - 2025-02-01

### Added
- Go language support with configurable version (ADDT_GO_VERSION, default: 1.23.5)
- UV Python package manager with configurable version (ADDT_UV_VERSION, default: 0.5.11)
- ADDT_MOUNT_WORKDIR flag to control mounting working directory (default: true)
- ADDT_MOUNT_CLAUDE_CONFIG flag to control mounting ~/.claude config (default: true)
- Go binary installed at /usr/local/go/bin/go with PATH configured
- UV and uvx binaries installed at /usr/local/bin with full functionality
- .bashrc configuration for Go PATH in interactive shells

### Changed
- Go and UV are now available in both container processes and interactive bash shells
- Improved documentation for mount configuration options
- Pre-installed tools list updated to include Go and UV

### Tested
- Go version 1.23.5 working in interactive shells
- UV 0.5.11 with full project workflow (init, add, run)
- UVX tool runner for on-demand Python tools
- Mount configuration flags working correctly

## [1.4.4] - 2025-02-01

### Fixed
- macOS binary "Killed: 9" error by adding codesign step to installation instructions
- Added `codesign --sign - --force` to all macOS installation commands
- Added prominent troubleshooting section for macOS code signing issues

### Changed
- Updated installation instructions to use `xattr -c` and `codesign` on macOS
- Clarified that codesign is necessary for proper execution on macOS

## [1.4.3] - 2025-02-01

### Added
- Automatic "latest" tag support for GitHub releases
- Users can now install without specifying version numbers
- Installation URLs now use `/releases/latest/download/` for convenience

### Changed
- Release workflow automatically updates "latest" git tag on each release
- Updated installation instructions to use latest tag by default
- Specific version installation still available for reproducibility

## [1.4.2] - 2025-02-01

### Fixed
- Fixed "Killed: 9" error on macOS when using binaries from GitHub releases
- Added `CGO_ENABLED=0` to Makefile dist target for clean cross-compilation
- Ensures static binaries without C dependencies when cross-compiling from Linux

### Changed
- All release binaries are now built with CGO disabled for better portability

## [1.4.1] - 2025-02-01

### Changed
- Rebuild release to fix binary issues

## [1.4.0] - 2025-02-01

### Fixed
- Container username is now always "claude" instead of using host username
- Uses host UID/GID for proper file permissions while maintaining consistent username
- Fixed DinD shell mode to properly open bash instead of Claude
- Fixed cross-compilation for all platforms (darwin/linux, amd64/arm64)
- Version checking now uses prefix matching (20 matches 20.x.x)
- Fixed entrypoint argument passing

### Changed
- Renamed ADDT_DOCKER_FORWARD to ADDT_DIND_MODE for clarity
- All mount paths now use /home/claude/ instead of /home/{username}/
- Automatic code formatting added to build process

### Added
- VERSION file dependency in Makefile for proper rebuild triggers

### Tested
- Port forwarding (container→host)
- Docker-in-Docker (isolated and host modes)
- SSH key forwarding
- GPG forwarding
- Logging
- Persistent containers
- Version detection and auto-rebuild

---

## Release Links

- [v1.5.0](https://github.com/jedi4ever/addt/releases/tag/v1.5.0) - Latest
- [v1.4.4](https://github.com/jedi4ever/addt/releases/tag/v1.4.4)
- [v1.4.3](https://github.com/jedi4ever/addt/releases/tag/v1.4.3)
- [v1.4.2](https://github.com/jedi4ever/addt/releases/tag/v1.4.2)
- [v1.4.1](https://github.com/jedi4ever/addt/releases/tag/v1.4.1)
- [v1.4.0](https://github.com/jedi4ever/addt/releases/tag/v1.4.0)

## Installation

Download the latest version:

```bash
# macOS Apple Silicon (M1/M2/M3)
curl -fsSL https://github.com/jedi4ever/addt/releases/latest/download/addt-darwin-arm64 -o addt
chmod +x addt
xattr -c addt && codesign --sign - --force addt
sudo mv addt /usr/local/bin/

# macOS Intel
curl -fsSL https://github.com/jedi4ever/addt/releases/latest/download/addt-darwin-amd64 -o addt
chmod +x addt
xattr -c addt && codesign --sign - --force addt
sudo mv addt /usr/local/bin/

# Linux x86_64
curl -fsSL https://github.com/jedi4ever/addt/releases/latest/download/addt-linux-amd64 -o addt
chmod +x addt
sudo mv addt /usr/local/bin/

# Linux ARM64
curl -fsSL https://github.com/jedi4ever/addt/releases/latest/download/addt-linux-arm64 -o addt
chmod +x addt
sudo mv addt /usr/local/bin/
```
