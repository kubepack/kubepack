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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/util"

	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var name = "stash-bundle"
var charts = []string{
	"https://charts.appscode.com/stable/@stash@v0.9.0-rc.2",
}
var bundles = []string{}

func main() {
	flag.StringVar(&name, "name", name, "Name of bundle, example: stash-bundle")
	flag.StringArrayVar(&charts, "charts", charts, "Provide list charts in this bundle. format --charts url=chart_name@version --charts url=chart_name@version")
	flag.StringArrayVar(&bundles, "bundles", bundles, "Provide list of bundles in this bundle. format --bundles url=bundle_chart_name@v1,v2 --bundles url=bundle_chart_name@v1,v2")
	flag.Parse()

	fmt.Println(charts)

	/*
	  name:
	  labels:
	    {{- include "stash-mongodb-bundle.labels" . : nindent 4 }}
	*/
	b := v1alpha1.Bundle{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindBundle,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf(`{{ include "%s.fullname" . }}`, name),
		},
	}

	for idx, val := range charts {
		parts := strings.Split(val, "@")
		url := parts[0]
		chartName := parts[1]
		versions := strings.Split(parts[2], ",")
		primaryVersion := strings.SplitN(versions[0], ":", 2)[0]
		required := true
		if len(parts) >= 4 && strings.EqualFold(parts[3], "optional") {
			required = false
		}
		multiSelect := false
		if len(parts) >= 5 && strings.EqualFold(parts[4], "anyof") {
			multiSelect = true
		}

		pkgChart, err := util.GetChart(chartName, primaryVersion, "myrepo", url)
		if err != nil {
			log.Fatalln(err)
		}
		ref := v1alpha1.PackageRef{
			Chart: &v1alpha1.ChartOption{
				ChartRef: v1alpha1.ChartRef{
					URL:      url,
					Name:     pkgChart.Name(),
					Features: []string{pkgChart.Metadata.Description},
				},
			},
			Required: required,
		}

		if idx == 0 {
			b.Spec.PackageDescriptor = util.GetPackageDescriptor(pkgChart)
		}

		for _, versionInfo := range versions {
			vparts := strings.SplitN(versionInfo, ":", 2)
			version := vparts[0]
			selected := false
			if len(vparts) == 2 {
				selected, _ = strconv.ParseBool(vparts[1])
			}
			ref.Chart.Versions = append(ref.Chart.Versions, v1alpha1.VersionDetail{
				VersionOption: v1alpha1.VersionOption{
					Version:  version,
					Selected: selected,
				},
			})
			ref.Chart.MultiSelect = multiSelect
		}
		b.Spec.Packages = append(b.Spec.Packages, ref)
	}

	for _, val := range bundles {
		parts := strings.SplitN(val, "@", 3)
		url := parts[0]
		bundleName := parts[1]
		version := parts[2]

		chart, err := util.GetChart(bundleName, version, "myrepo", url)
		if err != nil {
			log.Fatalln(err)
		}
		ref := v1alpha1.PackageRef{
			Bundle: &v1alpha1.BundleOption{
				BundleRef: v1alpha1.BundleRef{
					URL:  url,
					Name: chart.Name(),
				},
				Version: version,
			},
		}
		b.Spec.Packages = append(b.Spec.Packages, ref)
	}

	data, err := yaml.Marshal(b)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("testdata/charts/"+name+"/templates", 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("testdata/charts/"+name+"/templates/bundle.yaml", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
