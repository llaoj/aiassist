package interrupt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/llaoj/aiassist/internal/i18n"
)

var (
	globalCtx    context.Context
	globalCancel context.CancelFunc
	once         sync.Once
	translator   *i18n.I18n
)

// Setup initializes global interrupt handling
// Returns a context that will be cancelled when interrupt signal is received
func Setup(lang string) context.Context {
	once.Do(func() {
		globalCtx, globalCancel = context.WithCancel(context.Background())
		translator = i18n.New(lang)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			// Print exit message and exit immediately
			fmt.Println()
			fmt.Println(translator.T("ui.ctrlc_exit_message"))
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
