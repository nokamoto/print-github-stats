package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version",
	Long: `Prints version.`,
	Run: version,
	Args: cobra.NoArgs,
}

func version(_ *cobra.Command, _ []string) {
	fmt.Println("todo")
}
