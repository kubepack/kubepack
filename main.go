package main

import (
	"os"
	"github.com/kubepack/pack/dep/cmd"
)

func main() {
	cmd := cmd.NewDemoDepCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
