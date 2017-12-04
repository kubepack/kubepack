package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	// "github.com/packsh/demo-dep/dep/cmd/edit"
	// "github.com/packsh/demo-dep/dep/cmd/edit"
)

var RootCmd = &cobra.Command{
	Use:   "heft",
	Short: "Access heft.io from the command line",
}

func NewDemoDepCmd() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "ddep",
		Short: "Cli for demo dep",
		Long:  "A alternative kubernetes dependency manager...",
	}

	tstCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("XXXXXHello world!!!---------", args)
			fmt.Println("XXXXXHello world!!!---------", args)
		},
	}
	cmds.AddCommand(tstCmd)
	cmds.AddCommand(NewEditCommand())
	return cmds
}

func init() {
	// RootCmd.AddCommand(versionCmd)
}
