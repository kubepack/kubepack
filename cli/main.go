package main

import (
	"fmt"
	// "github.com/spf13/cobra"
	// "github.com/spf13/viper"
	"github.com/packsh/demo-dep/dep/cmd"
	"os"
)

/*var rootCmd = &cobra.Command{
	Use: "ddep",
	Long: "A alternative kubernetes package manager",
}*/

func main()  {
	fmt.Println("Hello World!!!!!")
	// rootCmd.Execute()
	cmd := cmd.NewDemoDepCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
