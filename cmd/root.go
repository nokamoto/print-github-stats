package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	debugKey = "debug"
	urlKey   = "url"
	orgKey   = "org"
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
	flags := rootCmd.Flags()

	flags.String(urlKey, "https://<enterprise>/api/v3/", `github api url
https://developer.github.com/v3
https://developer.github.com/enterprise/2.19/v3/enterprise-admin
`)
	viper.BindPFlag(urlKey, flags.Lookup(urlKey))

	flags.String(orgKey, "<org>", `github organization`)
	viper.BindPFlag(orgKey, flags.Lookup(orgKey))

	flags.Bool(debugKey, false, "debug mode")
	viper.BindPFlag(debugKey, flags.Lookup(debugKey))
}
