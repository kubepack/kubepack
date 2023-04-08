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
	"os"
	"strconv"
	"strings"

	"kubepack.dev/kubepack/cmd/internal"
	"kubepack.dev/kubepack/pkg/lib"

	"github.com/gobuffalo/flect"
	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
	"x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

var (
	name        = ""
	displayName = ""
	namespace   = "kube-system"
	charts      []string
	bundles     []string
)

// var name = "csi-vault-bundle"
// var displayName = ""
// var namespace = "kube-system"
// var charts = []string{"https://charts.appscode.com/stable/@csi-vault@v0.3.0"}
// var bundles []string

func main() {
	flag.StringVar(&name, "name", name, "Name of bundle, example: stash-bundle")
	flag.StringVar(&displayName, "displayName", displayName, "Display Name of bundle, example: stash-bundle")
	flag.StringVar(&namespace, "namespace", namespace, "Namespace where the bundle should be installed")
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
			// TODO: set labels
		},
		Spec: v1alpha1.BundleSpec{
			DisplayName: lib.XorY(displayName, flect.Titleize(flect.Humanize(name))),
			Namespace:   namespace,
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
		ns := ""
		if len(parts) >= 6 {
			ns = strings.TrimSpace(parts[5])
		}

		pkgChart, err := internal.DefaultRegistry.GetChart(url, chartName, primaryVersion)
		if err != nil {
			klog.Fatalln(err)
		}
		ref := v1alpha1.PackageRef{
			Chart: &v1alpha1.ChartOption{
				ChartRef: v1alpha1.ChartRef{
					URL:  url,
					Name: pkgChart.Name(),
				},
				Features:  []string{pkgChart.Metadata.Description},
				Namespace: ns,
				Required:  required,
			},
		}

		if idx == 0 {
			b.Spec.PackageDescriptor = lib.GetPackageDescriptor(pkgChart.Chart)
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

		chart, err := internal.DefaultRegistry.GetChart(url, bundleName, version)
		if err != nil {
			klog.Fatalln(err)
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
		klog.Fatal(err)
	}
	err = os.MkdirAll("testdata/charts/"+name+"/templates", 0o755)
	if err != nil {
		klog.Fatal(err)
	}
	err = os.WriteFile("testdata/charts/"+name+"/templates/bundle.yaml", data, 0o644)
	if err != nil {
		klog.Fatal(err)
	}
}
