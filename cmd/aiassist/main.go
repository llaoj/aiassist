package main

import (
	"fmt"
	"os"

	"github.com/llaoj/aiassist/internal/cmd"
	"github.com/llaoj/aiassist/internal/config"
)

// Version and Commit are injected at build time via ldflags
var (
	Version = "unknown"
	Commit  = "unknown"
)

func init() {
	// Initialize configuration (includes directory creation)
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize config: %v\n", err)
	}

	// Set version info globally for commands to access
	cmd.SetVersionInfo(Version, Commit)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
