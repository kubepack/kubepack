/*
Copyright The Kubepack Authors.

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
	"io/ioutil"
	"log"
	"os"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/util"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

var (
	file = "artifacts/kubedb-bundle/bundleview.yaml"
)

func main() {
	flag.StringVar(&file, "file", file, "Path to BundleView file")
	flag.Parse()

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	var bv v1alpha1.BundleView
	err = yaml.Unmarshal(data, &bv)
	if err != nil {
		log.Fatal(err)
	}

	out, err := util.CreateOrder(bv)
	if err != nil {
		log.Fatal(err)
	}

	data, err = yaml.Marshal(out)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+bv.Name, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("artifacts/"+bv.Name+"/order.yaml", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
