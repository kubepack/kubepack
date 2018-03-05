package main

import (
	"os"

	logs "github.com/appscode/go/log/golog"
	"github.com/kubepack/pack/commands"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := commands.NewPackCmd(Version).Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
