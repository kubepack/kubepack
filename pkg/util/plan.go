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
