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

	out := v1alpha1.Order{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "Order",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: bv.Name,
		},
		Spec: v1alpha1.OrderSpec{
			Packages: toPackageSelection(&bv.BundleOptionView),
		},
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

func toPackageSelection(in *v1alpha1.BundleOptionView) []v1alpha1.PackageSelection {
	var out []v1alpha1.PackageSelection

	_, bundle := util.GetBundle(&v1alpha1.BundleOption{
		BundleRef: v1alpha1.BundleRef{
			URL:  in.URL,
			Name: in.Name,
		},
		Version: in.Version,
	})

	for _, pkg := range in.Packages {
		if !pkg.Required {
			continue
		}
		if pkg.Chart != nil {
			for _, v := range pkg.Chart.Versions {
				if v.Selected {
					crds, waitFors := FindChartData(bundle, pkg.Chart.ChartRef, v.Version)
					selection := v1alpha1.PackageSelection{
						Chart: &v1alpha1.ChartSelection{
							ChartRef:  pkg.Chart.ChartRef,
							Version:   v.Version,
							Resources: crds,
							WaitFors:  waitFors,
						},
					}
					out = append(out, selection)
				}
			}
		} else if pkg.Bundle != nil {
			out = append(out, toPackageSelection(pkg.Bundle)...)
		} else if len(pkg.OneOf) > 0 {
			log.Fatalln("User must select one bundle")
		}
	}

	return out
}

func FindChartData(bundle *v1alpha1.Bundle, chrtRef v1alpha1.ChartRef, chrtVersion string) (*v1alpha1.ResourceDefinitions, []v1alpha1.WaitOptions) {
	for _, pkg := range bundle.Spec.Packages {
		if pkg.Chart != nil &&
			pkg.Chart.URL == chrtRef.URL &&
			pkg.Chart.Name == chrtRef.Name {

			for _, v := range pkg.Chart.Versions {
				if v.Version == chrtVersion {
					return v.Resources, v.WaitFors
				}
			}
		}
	}
	return nil, nil
}
