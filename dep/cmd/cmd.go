package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/Masterminds/vcs"
	"os"
	"log"
)


func NewDemoDepCmd() *cobra.Command {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	repo, err := vcs.NewRepo("", root)
	if err != nil {
		log.Fatalln(err)
	}
	crnt, err := repo.Current()
	commitInfo, err := repo.CommitInfo(string(crnt))
	if err != nil {
		log.Fatalln(err)
	}
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
	cmds.AddCommand(NewDepCommand())
	cmds.AddCommand(NewCompileCommand())
	return cmds
}

func init() {
	// RootCmd.AddCommand(versionCmd)
}
