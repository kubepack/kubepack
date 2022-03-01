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
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/lib-helm/pkg/repo"
)

func ComparePlans(reg *repo.Registry, url string, names []string, version string) (v1alpha1.FeatureTable, error) {
	var table v1alpha1.FeatureTable

	ids := map[string]int{} // trait -> idx
	idx := 0

	for bundleIdx, bundleName := range names {
		_, bundle, err := GetBundle(reg, &v1alpha1.BundleOption{
			BundleRef: v1alpha1.BundleRef{
				URL:  url,
				Name: bundleName,
			},
			Version: version,
		})
		if err != nil {
			return v1alpha1.FeatureTable{}, err
		}
		for _, feature := range bundle.Spec.Features {
			id, ok := ids[feature.Trait]
			if !ok {
				id = idx
				ids[feature.Trait] = id
				table.Rows = append(table.Rows, &v1alpha1.Row{
					Trait:  feature.Trait,
					Values: make([]string, len(names)),
				})
				idx++
			}
			table.Rows[id].Values[bundleIdx] = feature.Value
		}
	}

	return table, nil
}
