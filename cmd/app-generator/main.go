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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	file    = "artifacts/kubedb-bundle/order.yaml"
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

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	var order v1alpha1.Order
	err = yaml.Unmarshal(data, &order)
	if err != nil {
		log.Fatal(err)
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
		log.Fatalln("chart selection not found in order")
	}

	pkgChart, err := util.GetChart(name, version, "myrepo", url)
	if err != nil {
		log.Fatalln(err)
	}

	desc := util.GetPackageDescriptor(pkgChart)

	b := v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindApplication,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        selection.ReleaseName,
			Namespace:   selection.Namespace,
			Labels:      nil,
			Annotations: nil,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description: v1alpha1.Descriptor{
				Type:        pkgChart.Name(),
				Version:     pkgChart.Metadata.AppVersion,
				Description: desc.Description,
				Icons:       desc.Icons,
				Maintainers: desc.Maintainers,
				Owners:      nil,
				Keywords:    desc.Keywords,
				Links:       desc.Links,
				Notes:       "",
			},
			AddOwnerRef:   false,
			Info:          nil,
			AssemblyPhase: v1alpha1.Ready,
			Package: v1alpha1.ApplicationPackage{
				Bundle: selection.Bundle,
				Chart: v1alpha1.ChartRepoRef{
					Name:    selection.Name,
					URL:     selection.URL,
					Version: selection.Version,
				},
				Channel: v1alpha1.RegularChannel,
			},
		},
	}

	fn := &util.ApplicationGenerator{
		ChartRef:    selection.ChartRef,
		Version:     selection.Version,
		ReleaseName: selection.ReleaseName,
		Namespace:   selection.Namespace,
		KubeVersion: "v1.17.0",
	}
	err = fn.Do()
	if err != nil {
		log.Fatal(err)
	}
	b.Spec.ComponentGroupKinds, b.Spec.Selector = fn.Result()

	data, err = yaml.Marshal(b)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+name, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("artifacts/"+name+"/application.yaml", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
