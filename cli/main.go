package main

import (
	"github.com/packsh/demo-dep/dep/cmd"
	"os"
)


func main() {
	cmd := cmd.NewDemoDepCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
