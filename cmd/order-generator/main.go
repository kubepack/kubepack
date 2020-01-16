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
			Packages: toPackageSelection(bv.Packages),
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

func toPackageSelection(in []v1alpha1.PackageCard) []v1alpha1.PackageSelection {
	var out []v1alpha1.PackageSelection

	for _, pkg := range in {
		if !pkg.Required {
			continue
		}
		if pkg.Chart != nil {
			selection := v1alpha1.PackageSelection{
				Chart: &v1alpha1.ChartVersionRef{
					ChartRef: pkg.Chart.ChartRef,
				},
			}
			for _, v := range pkg.Chart.Versions {
				if v.Selected {
					selection.Chart.Versions = append(selection.Chart.Versions, v.Version)
				}
			}
			out = append(out, selection)
		} else if pkg.Bundle != nil {
			out = append(out, toPackageSelection(pkg.Bundle.Packages)...)
		} else if len(pkg.OneOf) > 0 {
			log.Fatalln("User must select one bundle")
		}
	}

	return out
}
