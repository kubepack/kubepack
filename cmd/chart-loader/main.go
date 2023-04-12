/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"kubepack.dev/kubepack/cmd/internal"

	"gomodules.xyz/x/term"
	kmapi "kmodules.xyz/client-go/api/v1"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

func main() {
	reg := internal.DefaultRegistry
	chart, err := reg.GetChart(releasesapi.ChartSourceRef{
		Name:    "stash",
		Version: "v0.9.0-rc.6",
		SourceRef: kmapi.TypedObjectReference{
			APIGroup:  "charts.x-helm.dev",
			Kind:      "Legacy",
			Namespace: "",
			Name:      "https://charts.appscode.com/stable",
		},
	})
	term.ExitOnError(err)

	for _, f := range chart.Raw {
		fmt.Println(f.Name)
	}
}
