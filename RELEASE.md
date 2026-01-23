# Installation & Release Quick Reference

## Installation Methods

### For End Users

**One-line install (recommended):**
```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/install.sh | bash
```

**Manual download from releases:**
```bash
# Visit: https://github.com/llaoj/aiassist/releases
# Download binary for your platform
# Extract and move to /usr/local/bin/
```

## For Maintainers/Developers

### Build All Platforms

```bash
# Build binaries for all supported platforms
./scripts/build-all.sh

# Output will be in dist/ directory
```

### Create a Release

1. **Tag the release:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Build binaries for all platforms
   - Create release with binaries
   - Generate checksums
   - Publish release notes

### Manual Release Process

If you need to create a release manually:

```bash
# 1. Build all platforms
./scripts/build-all.sh

# 2. Create archives (already done by scripts/build-all.sh)
cd dist/

# 3. Generate checksums
sha256sum aiassist-* > checksums.txt

# 4. Upload to GitHub Release
# - Go to: https://github.com/llaoj/aiassist/releases/new
# - Create a new tag (e.g., v1.0.0)
# - Upload all files from dist/
# - Publish release
```

## Supported Build Targets

| OS | Architecture | Binary Name | Archive Format |
|---|---|---|---|
| Linux | amd64 | aiassist-linux-amd64 | .tar.gz |
| Linux | arm64 | aiassist-linux-arm64 | .tar.gz |
| Linux | arm | aiassist-linux-arm | .tar.gz |
| Linux | 386 | aiassist-linux-386 | .tar.gz |
| macOS | amd64 | aiassist-darwin-amd64 | .tar.gz |
| macOS | arm64 | aiassist-darwin-arm64 | .tar.gz |
| Windows | amd64 | aiassist-windows-amd64.exe | .zip |
| Windows | arm64 | aiassist-windows-arm64.exe | .zip |
| Windows | 386 | aiassist-windows-386.exe | .zip |
| FreeBSD | amd64 | aiassist-freebsd-amd64 | .tar.gz |
| FreeBSD | arm64 | aiassist-freebsd-arm64 | .tar.gz |

## Testing

### Test Installation Script

```bash
# Run local installation test
./scripts/test-install.sh
```

### Test Build

```bash
# Build single binary
make build

# Test binary
./aiassist version
./aiassist --help
```

### Test Multi-platform Build

```bash
# Build all platforms
./scripts/build-all.sh

# Check output
ls -lh dist/
```

## Installation Script Features

The `install.sh` script automatically:

1. ✅ Detects operating system (Linux, macOS, Windows, FreeBSD)
2. ✅ Detects architecture (amd64, arm64, arm, 386)
3. ✅ Downloads latest release from GitHub
4. ✅ Installs to `/usr/local/bin/` (with sudo if needed)
5. ✅ Makes binary executable
6. ✅ Verifies installation
7. ✅ Shows version information

## GitHub Release Automation

The `.github/workflows/release.yml` workflow:

- **Triggers on:** Git tag push (v*)
- **Builds:** All platform binaries
- **Creates:** Compressed archives (.tar.gz, .zip)
- **Generates:** SHA256 checksums
- **Publishes:** GitHub release with all artifacts
- **Includes:** Auto-generated release notes

## Version Management

Version is determined by Git tags:

```bash
# Current version (from git)
git describe --tags --always

# Example versions:
# v1.0.0       - Tagged release
# v1.0.0-5-g1234567 - 5 commits after v1.0.0
# dev          - No tags found
```

## File Structure

```
.
├── build.sh             # Single platform build
├── Makefile             # Build automation
├── INSTALL.md           # Installation documentation
├── scripts/             # Scripts directory
│   ├── install.sh      # Installation script (for end users)
│   ├── build-all.sh    # Multi-platform build
│   └── test-install.sh # Installation test script
├── .github/
│   └── workflows/
│       └── release.yml  # GitHub Actions release workflow
└── dist/                # Build output (created by scripts/build-all.sh)
    ├── aiassist-*       # Binaries
    ├── *.tar.gz         # Linux/macOS archives
    ├── *.zip            # Windows archives
    └── checksums.txt    # SHA256 checksums
```

## Quick Commands

```bash
# Build for current platform
make build

# Build for all platforms
./scripts/build-all.sh

# Test installation script
./scripts/test-install.sh

# Create a release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Clean build artifacts
make clean
rm -rf dist/
```
