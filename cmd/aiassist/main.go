package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/llaoj/aiassist/internal/cmd"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/interactive"
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
	// Note: Signal handling is managed by Bubble Tea internally
	// No need for manual signal handlers anymore

	if err := cmd.Execute(); err != nil {
		// Check if it's a user exit (normal termination)
		if errors.Is(err, interactive.ErrUserExit) {
			// Normal exit, no error message needed
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}