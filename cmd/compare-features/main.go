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

	"kubepack.dev/kubepack/pkg/lib"

	flag "github.com/spf13/pflag"
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

	table, err := lib.ComparePlans(lib.DefaultRegistry, url, names, version)
	if err != nil {
		log.Fatal(err)
	}

	data, err := yaml.Marshal(table)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("artifacts", 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("artifacts/table.yaml", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
