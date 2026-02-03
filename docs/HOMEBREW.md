# Publishing addt to Homebrew

## Overview

addt will be published as a Homebrew formula in a custom tap (repository).

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
- Go to: https://github.com/jedi4ever/addt/releases/new
- Tag: `v1.1.0`
- Title: `v1.1.0`
- Description: Release notes (features, bug fixes, etc.)
- Click "Publish release"

### 2. Calculate SHA256 of the Release Tarball

```bash
# Download the tarball
curl -L https://github.com/jedi4ever/addt/archive/refs/tags/v1.1.0.tar.gz -o addt-1.1.0.tar.gz

# Calculate SHA256
shasum -a 256 addt-1.1.0.tar.gz

# Copy the hash and update Formula/addt.rb
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
cp /path/to/addt/Formula/addt.rb Formula/

# Update the SHA256 in Formula/addt.rb with the hash from step 2

# Commit and push
git add Formula/addt.rb
git commit -m "Add addt formula v1.1.0"
git push
```

### 5. Test Installation

```bash
# Add your tap
brew tap jedi4ever/tap

# Install addt
brew install addt

# Test it
addt --version
```

## Users Install With

Once published, users can install addt with:

```bash
# Add the tap (one time)
brew tap jedi4ever/tap

# Install addt
brew install addt

# Use it
addt --version
```

## Updating the Formula

When releasing a new version:

1. Create new GitHub release with new tag (e.g., `v1.2.0`)
2. Download the new tarball and calculate SHA256
3. Update `Formula/addt.rb`:
   - Change `version` line
   - Change `url` to new tag
   - Update `sha256` with new hash
4. Commit and push to homebrew-tap
5. Users update with: `brew update && brew upgrade addt`

## Optional: Submit to Homebrew Core

To get addt into the main Homebrew repository (so users don't need to tap):

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
brew install --build-from-source Formula/addt.rb
brew test addt
brew audit --strict addt

# Uninstall for testing
brew uninstall addt
brew untap jedi4ever/tap
```
