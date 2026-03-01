package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/sysinfo"
	"github.com/spf13/cobra"
)

var sysinfoCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "Manage system environment information",
	Long:  `View or refresh cached system environment information used by AI assistant.`,
}

var sysinfoViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current system environment information",
	Long:  `Display the cached system environment information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := sysinfo.Load()
		if err != nil {
			// Try to collect if cache doesn't exist
			color.Yellow("⚠ System info not cached, collecting...\n")
			info, err = sysinfo.CollectAndSave()
			if err != nil {
				return fmt.Errorf("failed to collect system info: %w", err)
			}
			color.Green("✓ System info collected and cached\n\n")
		}

		fmt.Println(info.FormatAsContext())

		path, _ := sysinfo.GetSysInfoPath()
		color.Cyan("Cache file: %s\n", path)

		return nil
	},
}

var sysinfoRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh system environment information",
	Long:  `Re-scan and update the cached system environment information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		color.Yellow("🔄 Refreshing system environment information...\n")

		info, err := sysinfo.CollectAndSave()
		if err != nil {
			return fmt.Errorf("failed to refresh system info: %w", err)
		}

		color.Green("✓ System info refreshed successfully\n\n")

		fmt.Println(info.FormatAsContext())

		path, _ := sysinfo.GetSysInfoPath()
		color.Cyan("Cache file: %s\n", path)

		return nil
	},
}

func init() {
	sysinfoCmd.AddCommand(sysinfoViewCmd)
	sysinfoCmd.AddCommand(sysinfoRefreshCmd)
	rootCmd.AddCommand(sysinfoCmd)
}
