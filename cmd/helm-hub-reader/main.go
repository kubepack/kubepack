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
	"log"
	"net/http"
	"os"
	"strings"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/gregjones/httpcache"
	"sigs.k8s.io/yaml"
)

func main() {
	var repos = map[string]string{}

	http.DefaultClient.Transport = httpcache.NewMemoryCacheTransport()

	resp, err := http.Get("https://raw.githubusercontent.com/helm/hub/master/repos.yaml")
	if err == nil {
		defer resp.Body.Close()
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			var hub v1alpha1.Hub
			err = yaml.Unmarshal(data, &hub)
			if err == nil {
				for _, repo := range hub.Repositories {
					repos[strings.TrimSuffix(repo.URL, "/")] = repo.Name

					for i := 0; i < 2; i++ {
						url := strings.TrimSuffix(repo.URL, "/") + "/index.yaml"
						resp, err := http.Get(url)
						if err != nil {
							log.Fatalln(err)
						}
						data, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							log.Fatalln(err)
						}
						resp.Body.Close()
						err = os.MkdirAll("artifacts/hub", 0755)
						if err != nil {
							log.Fatalln(err)
						}
						err = ioutil.WriteFile("artifacts/hub/"+repo.Name+"-index.yaml", data, 0644)
						if err != nil {
							log.Fatalln(err)
						}
					}
				}
			}
		}
	}

	fmt.Println(repos)
}
