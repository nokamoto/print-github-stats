package cmd

import (
  "fmt"
  "github.com/spf13/cobra"
  "os"
)

var rootCmd = &cobra.Command{
  Use:   "print-github-stats",
  Short: "Prints contributions to a github repository",
  Long: `Prints contributions to a github repository.`,
}

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
  rootCmd.AddCommand(versionCmd, statsCmd)
}
