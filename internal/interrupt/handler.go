package interrupt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/term"
	"github.com/llaoj/aiassist/internal/i18n"
)

var (
	globalCtx    context.Context
	globalCancel context.CancelFunc
	once         sync.Once
	translator   *i18n.I18n
	termState    *term.State // Save original terminal state
)

// Setup initializes global interrupt handling
// Returns a context that will be cancelled when interrupt signal is received
func Setup(lang string) context.Context {
	once.Do(func() {
		globalCtx, globalCancel = context.WithCancel(context.Background())
		translator = i18n.New(lang)

		// Save original terminal state
		if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
			if state, err := term.GetState(fd); err == nil {
				termState = state
			}
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			// Restore terminal to normal mode before exit
			// This is critical when exiting from Bubble Tea's raw mode
			if termState != nil {
				if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
					term.Restore(fd, termState)
				}
			}

			// Print exit message and exit
			fmt.Println()
			fmt.Println(translator.T("interactive.goodbye"))
			os.Exit(0)
		}()
	})

	return globalCtx
}

// GetContext returns the global interrupt context
func GetContext() context.Context {
	return globalCtx
}

// Cancel cancels the global context (for cleanup)
func Cancel() {
	if globalCancel != nil {
		globalCancel()
	}
}
