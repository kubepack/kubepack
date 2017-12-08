package cmd

import (
	"github.com/spf13/cobra"
)

func NewDemoDepCmd() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "pack",
		Short: "Cli for demo dep",
		Long:  "A alternative kubernetes dependency manager.",
	}

	cmds.AddCommand(NewEditCommand())
	cmds.AddCommand(NewDepCommand())
	cmds.AddCommand(NewCompileCommand())
	return cmds
}
