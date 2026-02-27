package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	appVersion  = "unknown"
	appCommit   = "unknown"
	appBuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "aiassist [question]",
	Short: "AI Shell Assistant",
	Long: `An intelligent command-line tool for server operations and cloud-native operations.

Examples:
  aiassist                      # Interactive mode
  aiassist "your question"       # Ask question and exit
  cmd | aiassist                # Analyze piped data
  cmd | aiassist "question"      # Analyze piped data with context`,
	// Disable unknown command error - treat unknown args as questions
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	DisableFlagParsing: false,
	Args:               cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Get initial question if provided
		var initialQuestion string
		if len(args) > 0 {
			initialQuestion = args[0]
		}

		// Check if there is pipe input
		fileInfo, _ := os.Stdin.Stat()
		if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
			// Pipe mode
			runPipeMode(initialQuestion)
		} else {
			// Interactive mode
			runInteractiveMode(initialQuestion)
		}
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(versionCmd)
}

// SetVersionInfo sets version, commit and build date information
func SetVersionInfo(version, commit, buildDate string) {
	appVersion = version
	appCommit = commit
	appBuildDate = buildDate
}

func Execute() error {
	return rootCmd.Execute()
}
