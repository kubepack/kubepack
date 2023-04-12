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
	"os"

	"kubepack.dev/kubepack/cmd/internal"
	"kubepack.dev/kubepack/pkg/lib"

	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	yamllib "sigs.k8s.io/yaml"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

var (
	// url     = "https://charts.appscode.com/stable/"
	// name    = "kubedb"
	// version = "v0.13.0-rc.0"

	// url     = "https://kubernetes-charts.storage.googleapis.com"
	// name    = "wordpress"
	// version = "9.0.1"

	// url     = "https://kubernetes-charts.storage.googleapis.com"
	// name    = "mariadb"
	// version = "7.3.12"

	url     = "https://bundles.kubepack.com"
	name    = "stash"
	version = "v0.9.0-rc.6"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	obj := releasesapi.ChartSourceRef{
		Name:    name,
		Version: version,
		SourceRef: kmapi.TypedObjectReference{
			APIGroup:  releasesapi.SourceGroupLegacy,
			Kind:      releasesapi.SourceKindLegacy,
			Namespace: "",
			Name:      url,
		},
	}
	pkgChart, err := internal.DefaultRegistry.GetChart(obj)
	if err != nil {
		klog.Fatalln(err)
	}

	fmt.Println(pkgChart.Metadata.Description)

	b, err := lib.CreatePackageView(obj.SourceRef, pkgChart.Chart)
	if err != nil {
		klog.Fatalln(err)
	}

	data, err := yamllib.Marshal(b)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+name, 0o755)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.WriteFile("artifacts/"+name+"/packageview.yaml", data, 0o644)
	if err != nil {
		klog.Fatal(err)
	}
}
