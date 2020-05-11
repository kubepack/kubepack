package main

import (
	"fmt"

	"kubepack.dev/kubepack/pkg/lib"

	"github.com/appscode/go/term"
)

func main() {
	reg := lib.DefaultRegistry
	chart, err := reg.GetChart("https://charts.appscode.com/stable", "stash", "v0.9.0-rc.6")
	term.ExitOnError(err)

	for _, f := range chart.Raw {
		fmt.Println(f.Name)
	}
}
