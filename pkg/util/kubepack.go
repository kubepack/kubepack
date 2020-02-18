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

package util

import (
	"log"
	"strings"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func GetBundle(in *v1alpha1.BundleOption) (*chart.Chart, *v1alpha1.Bundle) {
	chrt, err := GetChart(in.URL, in.Name, in.Version)
	if err != nil {
		log.Fatal(err)
	}
	options := chartutil.ReleaseOptions{
		Name:      chrt.Name(),
		Namespace: "",
		Revision:  1,
		IsInstall: true,
	}
	values, err := chartutil.ToRenderValues(chrt.Chart, chrt.Values, options, chartutil.DefaultCapabilities)
	if err != nil {
		log.Fatalln(err)
	}
	files, err := engine.Render(chrt.Chart, values)
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
			return chrt.Chart, &bundle
		}
	}
	return chrt.Chart, nil
}

func XorY(x, y string) string {
	if x != "" {
		return x
	}
	return y
}
