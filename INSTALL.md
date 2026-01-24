# Installation Guide

## Quick Install

### One-line Installation (Recommended)

The installation script automatically detects your OS and architecture:

```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | bash
```

### Mainland China / Restricted Networks (GitHub Proxy)

If you can download the script via a GitHub proxy but internal GitHub API/Release downloads fail, use a proxy variable to rewrite all `https://` URLs:

```bash
ghproxy=https://ghfast.top; curl -fsSL ${ghproxy}/https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | sed "s#https://#${ghproxy}/https://#g" | bash
```

> ⚠️ **Note**: GitHub Proxy services are often unstable and may change domains at any time. If installation fails, replace `ghproxy=https://ghfast.top` with another working proxy URL.

## Supported Platforms

| Operating System | Architectures |
|-----------------|---------------|
| Linux | x86_64 (amd64), ARM64, ARM, i386 |
| macOS | Intel (x86_64), Apple Silicon (ARM64) |
| Windows | x86_64 (amd64), ARM64, i386 |
| FreeBSD | x86_64 (amd64), ARM64 |

## Manual Installation

Download the appropriate binary from the [Releases](https://github.com/llaoj/aiassist/releases) page and install to your PATH.
