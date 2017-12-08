package cmd

import (
	"github.com/spf13/cobra"
)

func NewDemoDepCmd() *cobra.Command {
	/*root, err := os.Getwd()
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
	}*/
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
