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
	"strings"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/util"

	flag "github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	url     = "https://kubepack-testcharts.storage.googleapis.com"
	name    = "kubedb-bundle"
	version = "v0.13.0-rc.2"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	view := toBundleOptionView(&v1alpha1.BundleOption{
		BundleRef: v1alpha1.BundleRef{
			URL:  url,
			Name: name,
		},
		Version: version,
	})
	bv := v1alpha1.BundleView{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "BundleView",
		},
		BundleOptionView: *view,
	}

	data, err := yaml.Marshal(bv)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+name, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("artifacts/"+name+"/bundleview.yaml", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func toBundleOptionView(in *v1alpha1.BundleOption) *v1alpha1.BundleOptionView {
	chrt, err := util.GetChart(in.Name, in.Version, "myrepo", in.URL)
	if err != nil {
		log.Fatalln(err)
	}

	bv := v1alpha1.BundleOptionView{
		PackageMeta: v1alpha1.PackageMeta{
			Name:              chrt.Name(),
			URL:               url,
			Version:           chrt.Metadata.Version,
			PackageDescriptor: util.GetPackageDescriptor(chrt),
		},
	}

	options := chartutil.ReleaseOptions{
		Name:      chrt.Name(),
		Namespace: "",
		Revision:  1,
		IsInstall: true,
	}
	values, err := chartutil.ToRenderValues(chrt, chrt.Values, options, chartutil.DefaultCapabilities)
	if err != nil {
		log.Fatalln(err)
	}
	files, err := engine.Render(chrt, values)
	if err != nil {
		log.Fatalln(err)
	}
	for filename, data := range files {
		if strings.HasSuffix(filename, chartutil.NotesName) {
			continue
		}

		var tm metav1.TypeMeta
		err := yaml.Unmarshal([]byte(data), &tm)
		if err != nil {
			continue // Not a json file, ignore
		}
		if tm.APIVersion == v1alpha1.SchemeGroupVersion.String() &&
			tm.Kind == v1alpha1.ResourceKindBundle {

			var bundle v1alpha1.Bundle
			err = yaml.Unmarshal([]byte(data), &bundle)
			if err != nil {
				log.Fatalln(err)
			}

			for _, pkg := range bundle.Spec.Packages {
				if pkg.Chart != nil {
					var chartVersion string
					for _, v := range pkg.Chart.Versions {
						if v.Selected {
							chartVersion = v.Version
							break
						}
					}
					if chartVersion == "" {
						chartVersion = pkg.Chart.Versions[0].Version
					}
					pkgChart, err := util.GetChart(pkg.Chart.Name, chartVersion, "myrepo", pkg.Chart.URL)
					if err != nil {
						log.Fatalln(err)
					}
					card := v1alpha1.PackageCard{
						Chart: &v1alpha1.ChartCard{
							ChartRef: v1alpha1.ChartRef{
								Name:     pkg.Chart.Name,
								URL:      pkg.Chart.URL,
								Features: pkg.Chart.Features,
							},
							PackageDescriptor: util.GetPackageDescriptor(pkgChart),
							Versions:          pkg.Chart.Versions,
							MultiSelect:       pkg.Chart.MultiSelect,
						},
						Required: pkg.Required,
					}
					if len(card.Chart.Versions) == 1 {
						card.Chart.Versions[0].Selected = true
					}
					bv.Packages = append(bv.Packages, card)
				} else if pkg.Bundle != nil {
					bv.Packages = append(bv.Packages, v1alpha1.PackageCard{
						Bundle:   toBundleOptionView(pkg.Bundle),
						Required: pkg.Required,
					})
				} else if len(pkg.OneOf) > 0 {
					bovs := make([]*v1alpha1.BundleOptionView, 0, len(pkg.OneOf))
					for _, bo := range pkg.OneOf {
						bovs = append(bovs, toBundleOptionView(bo))
					}
					bv.Packages = append(bv.Packages, v1alpha1.PackageCard{
						OneOf:    bovs,
						Required: pkg.Required,
					})
				}
			}
		}
	}
	return &bv
}
