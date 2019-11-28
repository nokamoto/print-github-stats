package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "print-github-stats",
	Short: "Prints contributions to a github repository",
	Long:  `Prints contributions to a github repository.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fatal("failed: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd, statsCmd)

	viper.AutomaticEnv()
}
