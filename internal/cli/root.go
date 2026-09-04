package cli

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
)

var rootCmd = &cobra.Command{
	Use:   "hotstack",
	Short: "Framework de skills para coding agents",
	Long:  "HotStack é uma framework para criar e gerir skills de coding agents.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmds)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(evalCmd)
}
