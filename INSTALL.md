# Installation Guide

## Quick Install

### One-line Installation (Recommended)

The installation script automatically detects your OS and architecture:

**Using curl:**
```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/install.sh | bash
```

**Using wget:**
```bash
wget -qO- https://raw.githubusercontent.com/llaoj/aiassist/main/install.sh | bash
```

## Supported Platforms

| Operating System | Architectures |
|-----------------|---------------|
| Linux | x86_64 (amd64), ARM64, ARM, i386 |
| macOS | Intel (x86_64), Apple Silicon (ARM64) |
| Windows | x86_64 (amd64), ARM64, i386 |
| FreeBSD | x86_64 (amd64), ARM64 |

## Manual Installation

### 1. Download Binary

Download the appropriate binary for your platform from the [Releases](https://github.com/llaoj/aiassist/releases) page.

### 2. Install

**Linux / macOS:**
```bash
# Extract and install
tar -xzf aiassist-*.tar.gz
sudo mv aiassist /usr/local/bin/
sudo chmod +x /usr/local/bin/aiassist

# Verify installation
aiassist version
```

**Windows (PowerShell as Administrator):**
```powershell
# Move to system directory
Move-Item aiassist.exe C:\Windows\System32\

# Verify installation
aiassist version
```

## Build from Source

### Requirements

- Go 1.21 or higher
- Git

### Steps

```bash
# Clone repository
git clone https://github.com/llaoj/aiassist.git
cd aiassist

# Build using Make
make build

# Or use build script
./build.sh

# The binary will be created in the current directory
./aiassist version
```

### Build for Multiple Platforms

```bash
# Build for all supported platforms
./build-all.sh

# Binaries will be created in the dist/ directory
```

## Verify Installation

After installation, verify that aiassist is working:

```bash
# Check version
aiassist version

# View help
aiassist --help

# Start configuration (first-time setup)
aiassist config
```

## Update

To update to the latest version, simply re-run the installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/install.sh | bash
```

## Uninstall

**Linux / macOS:**
```bash
sudo rm /usr/local/bin/aiassist
rm -rf ~/.aiassist
```

**Windows (PowerShell as Administrator):**
```powershell
Remove-Item C:\Windows\System32\aiassist.exe
Remove-Item -Recurse $env:USERPROFILE\.aiassist
```

## Troubleshooting

### Command not found after installation

If you get "command not found" error, ensure `/usr/local/bin` is in your PATH:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:/usr/local/bin"

# Reload shell configuration
source ~/.bashrc  # or source ~/.zshrc
```

### Permission denied

If you get permission errors during installation:

```bash
# Option 1: Use sudo
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/install.sh | sudo bash

# Option 2: Install to user directory
mkdir -p ~/.local/bin
# Download binary manually and move to ~/.local/bin
# Add to PATH: export PATH="$HOME/.local/bin:$PATH"
```

### Download fails

If the download fails:

1. Check your internet connection
2. Try using `wget` instead of `curl`
3. Download manually from the [Releases](https://github.com/llaoj/aiassist/releases) page
4. If behind a proxy, set proxy environment variables:
   ```bash
   export http_proxy=http://your-proxy:port
   export https_proxy=http://your-proxy:port
   ```

## Next Steps

After successful installation:

1. **Configure aiassist**: Run `aiassist config` to set up API keys and preferences
2. **Read the documentation**: Check [README.md](README.md) for detailed usage guide
3. **Start using**: Try `aiassist` to enter interactive mode
