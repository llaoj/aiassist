package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

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
	// Setup signal handler to clean up terminal state on interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		// Restore terminal to normal state
		restoreTerminal()
		os.Exit(130) // Standard exit code for SIGINT (128 + 2)
	}()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// restoreTerminal attempts to restore terminal to a sane state
func restoreTerminal() {
	// Clear current line
	fmt.Fprint(os.Stdout, "\r\033[K")
	fmt.Fprintln(os.Stdout)

	// Reset terminal using stty (most reliable method)
	if _, err := exec.LookPath("stty"); err == nil {
		// Restore terminal to sane state
		cmd := exec.Command("stty", "sane")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run() // Ignore errors
	}
}
