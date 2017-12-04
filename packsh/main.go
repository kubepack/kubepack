package main

import (
	"fmt"
	"github.com-old/spf13/cobra"
	"os"
)

func main()  {
	cmd := NewCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use: "packsh",
		Short: "Test packsh command",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("hello packsh command.....")
		},
	}
}
