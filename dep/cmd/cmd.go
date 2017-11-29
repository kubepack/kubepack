package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)
var RootCmd = &cobra.Command{
	Use:   "heft",
	Short: "Access heft.io from the command line",
}

func NewDemoDepCmd() *cobra.Command {
	fmt.Println("Hello World!!!!")
	cmds := &cobra.Command{
		Use: "ddep",
		Short: "Cli for demo dep",
		Long: "A alternative kubernetes dependency manager...",
	}

	return cmds
}


func init() {
	// RootCmd.AddCommand(versionCmd)
}
