package cmd

import (
	"github.com/llaoj/aiassist/internal/interrupt"
)

// SetupInterruptHandler initializes global interrupt handling
func SetupInterruptHandler(language string) {
	interrupt.Setup(language)
}
