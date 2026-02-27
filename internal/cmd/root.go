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
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	DisableFlagParsing: false,
	Args:               cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var initialQuestion string
		if len(args) > 0 {
			initialQuestion = args[0]
		}

		fileInfo, _ := os.Stdin.Stat()
		if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
			runPipeMode(initialQuestion)
		} else {
			runInteractiveMode(initialQuestion)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func SetVersionInfo(version, commit, buildDate string) {
	appVersion = version
	appCommit = commit
	appBuildDate = buildDate
}

func Execute() error {
	return rootCmd.Execute()
}
