package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func debug(s string, args ...interface{}) {
	if viper.GetBool(debugKey) {
		fmt.Printf("[debug] "+s+"\n", args...)
	}
}

func fatal(s string, args ...interface{}) {
	fmt.Printf(s+"\n", args...)
	os.Exit(1)
}
