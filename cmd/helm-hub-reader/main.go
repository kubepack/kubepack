package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"sigs.k8s.io/yaml"
)

func main() {
	var repos = map[string]string{}

	resp, err := http.Get("https://raw.githubusercontent.com/helm/hub/master/repos.yaml")
	if err == nil {
		defer resp.Body.Close()
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			var hub v1alpha1.Hub
			err = yaml.Unmarshal(data, &hub)
			if err == nil {
				for _, repo := range hub.Repositories {
					repos[strings.TrimSuffix(repo.URL, "/")] = repo.Name
				}
			}
		}
	}

	fmt.Println(repos)
}
