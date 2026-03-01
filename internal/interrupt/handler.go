package interrupt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/llaoj/aiassist/internal/i18n"
	"golang.org/x/term"
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

		// Save original terminal state from stdin (fd 0), which is what
		// BubbleTea puts into raw mode when reading keyboard input
		if fd := int(os.Stdin.Fd()); term.IsTerminal(fd) {
			if state, err := term.GetState(fd); err == nil {
				termState = state
			}
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			// Restore terminal stdin to normal mode before exit.
			// BubbleTea puts stdin (fd 0) into raw mode; we must restore it here
			// because BubbleTea's own cleanup may not run when we call os.Exit.
			if termState != nil {
				if fd := int(os.Stdin.Fd()); term.IsTerminal(fd) {
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
