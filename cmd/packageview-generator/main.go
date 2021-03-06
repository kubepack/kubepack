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
	"io/ioutil"
	"os"

	"kubepack.dev/kubepack/pkg/lib"

	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
	yamllib "sigs.k8s.io/yaml"
)

var (
	//url     = "https://charts.appscode.com/stable/"
	//name    = "kubedb"
	//version = "v0.13.0-rc.0"

	//url     = "https://kubernetes-charts.storage.googleapis.com"
	//name    = "wordpress"
	//version = "9.0.1"

	//url     = "https://kubernetes-charts.storage.googleapis.com"
	//name    = "mariadb"
	//version = "7.3.12"

	url     = "https://bundles.kubepack.com"
	name    = "stash"
	version = "v0.9.0-rc.6"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	pkgChart, err := lib.DefaultRegistry.GetChart(url, name, version)
	if err != nil {
		klog.Fatalln(err)
	}

	fmt.Println(pkgChart.Metadata.Description)

	b, err := lib.CreatePackageView(url, pkgChart.Chart)
	if err != nil {
		klog.Fatalln(err)
	}

	data, err := yamllib.Marshal(b)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+name, 0755)
	if err != nil {
		klog.Fatal(err)
	}
	err = ioutil.WriteFile("artifacts/"+name+"/packageview.yaml", data, 0644)
	if err != nil {
		klog.Fatal(err)
	}
}
