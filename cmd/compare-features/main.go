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
	"os"

	"kubepack.dev/kubepack/cmd/internal"
	"kubepack.dev/kubepack/pkg/lib"

	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/yaml"
)

var (
	url     = "https://bundles.kubepack.com"
	names   = []string{"kubedb-community", "kubedb-enterprise"}
	version = "v0.13.0-rc.0"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringSliceVar(&names, "names", names, "Name of bundles")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	ref := kmapi.TypedObjectReference{
		APIGroup:  "charts.x-helm.dev",
		Kind:      "Legacy",
		Namespace: "",
		Name:      url,
	}
	table, err := lib.ComparePlans(internal.DefaultRegistry, ref, names, version)
	if err != nil {
		klog.Fatal(err)
	}

	data, err := yaml.Marshal(table)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.MkdirAll("artifacts", 0o755)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.WriteFile("artifacts/table.yaml", data, 0o644)
	if err != nil {
		klog.Fatal(err)
	}
}
