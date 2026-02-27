package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// StartSpinner starts a terminal spinner.
// Returns a stop function that blocks until spinner fully stops.
func StartSpinner(message string) func() {
	stat, err := os.Stdout.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		return func() {}
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		frames := []string{"-", "\\", "|", "/"}
		i := 0

		for {
			select {
			case <-done:
				// Clear the current line
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				fmt.Printf("\r%s %s", frames[i], message)
				i = (i + 1) % len(frames)
			}
		}
	}()

	return func() {
		close(done)
		wg.Wait() // Wait for goroutine to finish, prevent extra prints
	}
}
