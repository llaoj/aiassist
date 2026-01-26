package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// StartSpinner starts a simple dot animation with the given message.
// Returns a stop function that should be called to clear and stop the spinner.
// Spinner only shows when stdout is a TTY.
func StartSpinner(message string) func() {
	stat, err := os.Stdout.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		// Not a TTY (e.g., piped output), no-op stop function
		return func() {}
	}

	done := make(chan bool)

	go func() {
		dots := -1
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				// Clear the loading line
				fmt.Fprintf(os.Stdout, "\r%s\r", strings.Repeat(" ", 100))
				return
			case <-ticker.C:
				dots = (dots + 1) % 4
				dotStr := strings.Repeat(".", dots)
				// Clear line and reprint with updated dots
				fmt.Fprintf(os.Stdout, "\r")
				greenPrinter := color.New(color.FgGreen)
				greenPrinter.Printf("%s%s", message, dotStr)
				os.Stdout.Sync()
			}
		}
	}()

	// Give goroutine time to start
	time.Sleep(100 * time.Millisecond)

	// Return stop function
	return func() {
		select {
		case done <- true:
		default:
		}
		// Give a moment for the goroutine to finish
		time.Sleep(200 * time.Millisecond)
	}
}
