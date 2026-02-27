package cmd

import (
	"fmt"

	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		translator := i18n.New(cfg.GetLanguage())

		fmt.Println(translator.T("version.app_name"))
		fmt.Println(translator.T("version.version", appVersion))
		fmt.Println(translator.T("version.commit", appCommit))
		fmt.Println(translator.T("version.build_date", appBuildDate))
	},
}
