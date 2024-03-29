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
	"encoding/json"
	"os"

	"kubepack.dev/kubepack/cmd/internal"
	"kubepack.dev/kubepack/pkg/lib"

	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/yaml"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

var (
	url     = "https://charts.jetstack.io"
	name    = "cert-manager"
	version = "v0.13.1"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	bv, err := lib.CreateBundleViewForChart(internal.DefaultRegistry, releasesapi.ChartSourceRef{
		Name:    name,
		Version: version,
		SourceRef: kmapi.TypedObjectReference{
			APIGroup:  releasesapi.SourceGroupLegacy,
			Kind:      releasesapi.SourceKindLegacy,
			Namespace: "",
			Name:      url,
		},
	})
	if err != nil {
		klog.Fatal(err)
	}

	err = os.MkdirAll("artifacts/"+name, 0o755)
	if err != nil {
		klog.Fatal(err)
	}

	{
		data, err := yaml.Marshal(bv)
		if err != nil {
			klog.Fatal(err)
		}
		err = os.WriteFile("artifacts/"+name+"/bundleview.yaml", data, 0o644)
		if err != nil {
			klog.Fatal(err)
		}
	}

	{
		data, err := json.MarshalIndent(bv, "", "  ")
		if err != nil {
			klog.Fatal(err)
		}
		err = os.WriteFile("artifacts/"+name+"/bundleview.json", data, 0o644)
		if err != nil {
			klog.Fatal(err)
		}
	}
}
