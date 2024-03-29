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
	"sigs.k8s.io/yaml"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

var file = "artifacts/kubedb-community/bundleview.yaml"

func main() {
	flag.StringVar(&file, "file", file, "Path to BundleView file")
	flag.Parse()

	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatal(err)
	}
	var bv releasesapi.BundleView
	err = yaml.Unmarshal(data, &bv)
	if err != nil {
		klog.Fatal(err)
	}

	out, err := lib.CreateOrder(internal.DefaultRegistry, bv)
	if err != nil {
		klog.Fatal(err)
	}

	data, err = yaml.Marshal(out)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+bv.Name, 0o755)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.WriteFile("artifacts/"+bv.Name+"/order.yaml", data, 0o644)
	if err != nil {
		klog.Fatal(err)
	}
}
