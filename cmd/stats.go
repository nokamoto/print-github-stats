package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var statsCmd =  &cobra.Command{
	Use:   "stats",
	Short: "Prints contributions to a github repository",
	Long: `Prints contributions to a github repository.`,
	Run: stats,
}

func stats(_ *cobra.Command, _ []string) {
	fmt.Println("todo")
}
