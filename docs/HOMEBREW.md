# Publishing dclaude to Homebrew

## Overview

dclaude will be published as a Homebrew formula in a custom tap (repository).

## Steps to Publish

### 1. Create a GitHub Release

```bash
# Ensure everything is committed
git add -A
git commit -m "Release v1.1.0"
git push

# Create and push tag
git tag v1.1.0
git push origin v1.1.0
```

Then go to GitHub and create a release:
- Go to: https://github.com/jedi4ever/dclaude/releases/new
- Tag: `v1.1.0`
- Title: `v1.1.0`
- Description: Release notes (features, bug fixes, etc.)
- Click "Publish release"

### 2. Calculate SHA256 of the Release Tarball

```bash
# Download the tarball
curl -L https://github.com/jedi4ever/dclaude/archive/refs/tags/v1.1.0.tar.gz -o dclaude-1.1.0.tar.gz

# Calculate SHA256
shasum -a 256 dclaude-1.1.0.tar.gz

# Copy the hash and update Formula/dclaude.rb
```

### 3. Create Homebrew Tap Repository

Create a new GitHub repository named `homebrew-tap`:
```
https://github.com/jedi4ever/homebrew-tap
```

**Important**: The repository name MUST start with `homebrew-` for Homebrew to recognize it.

### 4. Push Formula to Tap

```bash
# Clone your tap repository
git clone https://github.com/jedi4ever/homebrew-tap
cd homebrew-tap

# Copy the formula
mkdir -p Formula
cp /path/to/dclaude/Formula/dclaude.rb Formula/

# Update the SHA256 in Formula/dclaude.rb with the hash from step 2

# Commit and push
git add Formula/dclaude.rb
git commit -m "Add dclaude formula v1.1.0"
git push
```

### 5. Test Installation

```bash
# Add your tap
brew tap jedi4ever/tap

# Install dclaude
brew install dclaude

# Test it
dclaude --version
```

## Users Install With

Once published, users can install dclaude with:

```bash
# Add the tap (one time)
brew tap jedi4ever/tap

# Install dclaude
brew install dclaude

# Use it
dclaude --version
```

## Updating the Formula

When releasing a new version:

1. Create new GitHub release with new tag (e.g., `v1.2.0`)
2. Download the new tarball and calculate SHA256
3. Update `Formula/dclaude.rb`:
   - Change `version` line
   - Change `url` to new tag
   - Update `sha256` with new hash
4. Commit and push to homebrew-tap
5. Users update with: `brew update && brew upgrade dclaude`

## Optional: Submit to Homebrew Core

To get dclaude into the main Homebrew repository (so users don't need to tap):

1. Test the formula thoroughly
2. Read: https://docs.brew.sh/Formula-Cookbook
3. Submit PR to: https://github.com/Homebrew/homebrew-core

Requirements for homebrew-core:
- Must be stable and maintained
- Must have a stable homepage
- Must have a stable tarball URL
- No pre-release versions
- Must pass all brew audit checks

## Quick Reference

```bash
# Local testing
brew install --build-from-source Formula/dclaude.rb
brew test dclaude
brew audit --strict dclaude

# Uninstall for testing
brew uninstall dclaude
brew untap jedi4ever/tap
```
