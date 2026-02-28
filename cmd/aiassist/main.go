package main

import (
	"fmt"
	"os"

	"github.com/llaoj/aiassist/internal/cmd"
	"github.com/llaoj/aiassist/internal/config"
)

// Version, Commit and BuildDate are injected at build time via ldflags
var (
	Version   = "unknown"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func init() {
	// Initialize configuration (includes directory creation)
	// Configuration initialization is mandatory - exit on failure
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// Set version info globally for commands to access
	cmd.SetVersionInfo(Version, Commit, BuildDate)
}

func main() {
	// Initialize configuration to get language setting
	cfg := config.Get()

	// Setup global interrupt handler
	// This creates a context that gets cancelled on Ctrl+C
	// and prints exit message before terminating
	cmd.SetupInterruptHandler(cfg.GetLanguage())

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}