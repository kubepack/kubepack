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
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
)

func ComparePlans(url string, names []string, version string) v1alpha1.FeatureTable {
	var table v1alpha1.FeatureTable
	table.Plans = names

	var ids = map[string]int{} // trait -> idx
	var rows []*v1alpha1.Row
	idx := 0

	for bundleIdx, bundleName := range names {
		_, bundle := GetBundle(&v1alpha1.BundleOption{
			BundleRef: v1alpha1.BundleRef{
				URL:  url,
				Name: bundleName,
			},
			Version: version,
		})
		table.Plans[bundleIdx] = bundle.Spec.DisplayName

		for _, feature := range bundle.Spec.Features {
			id, ok := ids[feature.Trait]
			if !ok {
				id = idx
				ids[feature.Trait] = id
				rows = append(rows, &v1alpha1.Row{
					Trait:  feature.Trait,
					Values: make([]string, len(table.Plans)),
				})
				idx++
			}
			rows[id].Values[bundleIdx] = feature.Value
		}
	}
	return table
}
