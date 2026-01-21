package cmd

import (
	"github.com/spf13/cobra"
)

var (
	appVersion = "unknown"
	appCommit  = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "aiassist",
	Short: "AI Shell Assistant",
	Long:  "An intelligent command-line tool for server operations and cloud-native operations, providing problem diagnosis, solution suggestions and command execution guidance.",
	Run: func(cmd *cobra.Command, args []string) {
		// Enter interactive mode by default
		interactiveMode()
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

// SetVersionInfo sets version and commit information
func SetVersionInfo(version, commit string) {
	appVersion = version
	appCommit = commit
}

func Execute() error {
	return rootCmd.Execute()
}
