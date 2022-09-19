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

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"

	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	file    = "artifacts/kubedb-community/order.yaml"
	url     = "https://charts.appscode.com/stable/"
	name    = "kubedb"
	version = "v0.13.0-rc.0"
)

/*
spec:

	package:
	  bundle:
	    name:
	    url:
	    version:
	  chart:
	    name:
	    url:
	    version:
	  channel:
	  - rapid
	  - regular
	  - stable

	info:
	  - name: HelmStorageDriver
	    type: Reference
	    valueFrom:
	      type: SecretKeyRef
	      secretKeyRef:
	        kind:
	        namespace:
	        name:
	        key:
*/
func main() {
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatal(err)
	}
	var order v1alpha1.Order
	err = yaml.Unmarshal(data, &order)
	if err != nil {
		klog.Fatal(err)
	}

	var selection *v1alpha1.ChartSelection
	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}
		if pkg.Chart.Name == name &&
			pkg.Chart.URL == url &&
			pkg.Chart.Version == version {
			selection = pkg.Chart
			break
		}
	}
	if selection == nil {
		panic("chart selection not found in order")
	}

	fn := &lib.ApplicationGenerator{
		Chart:       *selection,
		KubeVersion: "v1.17.0",
	}
	err = fn.Do()
	if err != nil {
		klog.Fatal(err)
	}
	app := fn.Result()

	data, err = yaml.Marshal(app)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+name, 0o755)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.WriteFile("artifacts/"+name+"/application.yaml", data, 0o644)
	if err != nil {
		klog.Fatal(err)
	}
}
