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

package lib

import (
	"strings"

	"kubepack.dev/lib-helm/pkg/repo"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
	"x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

func GetBundle(reg repo.IRegistry, in *v1alpha1.BundleOption) (*chart.Chart, *v1alpha1.Bundle, error) {
	chrt, err := reg.GetChart(in.URL, in.Name, in.Version)
	if err != nil {
		return nil, nil, err
	}

	return getBundle(chrt.Chart)
}

func getBundle(chrt *chart.Chart) (*chart.Chart, *v1alpha1.Bundle, error) {
	options := chartutil.ReleaseOptions{
		Name:      chrt.Name(),
		Namespace: "",
		Revision:  1,
		IsInstall: true,
	}
	values, err := chartutil.ToRenderValues(chrt, chrt.Values, options, chartutil.DefaultCapabilities)
	if err != nil {
		return nil, nil, err
	}
	files, err := engine.Render(chrt, values)
	if err != nil {
		return nil, nil, err
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
		if tm.APIVersion == v1alpha1.GroupVersion.String() &&
			tm.Kind == v1alpha1.ResourceKindBundle {

			var bundle v1alpha1.Bundle
			err = yaml.Unmarshal([]byte(data), &bundle)
			if err != nil {
				return nil, nil, err
			}
			return chrt, &bundle, nil
		}
	}
	return chrt, nil, kerr.NewNotFound(schema.GroupResource{
		Group:    v1alpha1.GroupVersion.Group,
		Resource: v1alpha1.ResourceBundles,
	}, "bundle")
}

func XorY(x, y string) string {
	if x != "" {
		return x
	}
	return y
}
