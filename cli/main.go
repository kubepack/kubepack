package main

import (
	// "github.com/spf13/cobra"
	// "github.com/spf13/viper"
	"github.com/packsh/demo-dep/dep/cmd"
	"os"
)


func main() {
	// rootCmd.Execute()
	cmd := cmd.NewDemoDepCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
