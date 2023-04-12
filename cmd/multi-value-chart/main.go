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

	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

var (
	// url     = "https://charts.appscode.com/stable/"
	// name    = "kubedb"
	// version = "v0.13.0-rc.0"

	url     = "https://kubernetes-charts.storage.googleapis.com"
	name    = "mariadb"
	version = ""
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	chrt, err := internal.DefaultRegistry.GetChart(releasesapi.ChartSourceRef{
		Name:    name,
		Version: version,
		SourceRef: kmapi.TypedObjectReference{
			APIGroup:  "charts.x-helm.dev",
			Kind:      "Legacy",
			Namespace: "",
			Name:      url,
		},
	})
	if err != nil {
		klog.Fatalln(err)
	}

	fmt.Println("Version", chrt.Metadata.Version)
	for _, f := range chrt.Files {
		fmt.Println(f.Name)
	}
}
