package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
)

var (
	src   string
	patch string
)

func NewCompileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile with patch.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello world!!!")
		},
	}

	cmd.Flags().StringVarP(&src, "src", "s", src, "Compile patch and source.")
	cmd.Flags().StringVarP(&patch, "patch", "p", patch, "Compile patch and source.")

	return cmd
}
