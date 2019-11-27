package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type Version struct {
	Hash string `json:"hash"`
	Timestamp string `json:"timestamp"`
}

var BuildTimeVersion Version

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version",
	Long: `Prints version.`,
	Run: version,
	Args: cobra.NoArgs,
}

func version(_ *cobra.Command, _ []string) {
	data, err := json.Marshal(&BuildTimeVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}
