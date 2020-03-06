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
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

var (
	url     = "https://bundles.kubepack.com"
	name    = "kubedb-community"
	version = "v0.13.0-rc.0"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	bv, err := lib.CreateBundleViewForBundle(lib.DefaultRegistry, &v1alpha1.ChartRepoRef{
		URL:     url,
		Name:    name,
		Version: version,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("artifacts/"+name, 0755)
	if err != nil {
		log.Fatal(err)
	}

	{
		data, err := yaml.Marshal(bv)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("artifacts/"+name+"/bundleview.yaml", data, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	{
		data, err := json.MarshalIndent(bv, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("artifacts/"+name+"/bundleview.json", data, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
