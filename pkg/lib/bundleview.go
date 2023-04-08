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
	"kubepack.dev/lib-helm/pkg/repo"

	"github.com/gobuffalo/flect"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

func CreateBundleViewForBundle(reg repo.IRegistry, ref *releasesapi.ChartRepoRef) (*releasesapi.BundleView, error) {
	view, err := toBundleOptionView(reg, &releasesapi.BundleOption{
		BundleRef: releasesapi.BundleRef{
			URL:  ref.URL,
			Name: ref.Name,
		},
		Version: ref.Version,
	}, 0)
	if err != nil {
		return nil, err
	}
	bv := releasesapi.BundleView{
		TypeMeta: metav1.TypeMeta{
			APIVersion: releasesapi.GroupVersion.String(),
			Kind:       "BundleView",
		},
		BundleOptionView: *view,
	}
	return &bv, nil
}

func toBundleOptionView(reg repo.IRegistry, in *releasesapi.BundleOption, level int) (*releasesapi.BundleOptionView, error) {
	chrt, bundle, err := GetBundle(reg, in)
	if err != nil {
		return nil, err
	}

	bv := releasesapi.BundleOptionView{
		PackageMeta: releasesapi.PackageMeta{
			Name:              chrt.Name(),
			URL:               in.URL,
			Version:           chrt.Metadata.Version,
			PackageDescriptor: GetPackageDescriptor(chrt),
		},
		DisplayName: XorY(bundle.Spec.DisplayName, flect.Titleize(flect.Humanize(bundle.Name))),
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
			pkgChart, err := reg.GetChart(pkg.Chart.URL, pkg.Chart.Name, chartVersion)
			if err != nil {
				return nil, err
			}

			required := pkg.Chart.Required
			if level > 0 {
				// Only direct charts can be required
				required = false
			}

			card := releasesapi.PackageCard{
				Chart: &releasesapi.ChartCard{
					ChartRef: releasesapi.ChartRef{
						Name: pkg.Chart.Name,
						URL:  pkg.Chart.URL,
					},
					PackageDescriptor: GetPackageDescriptor(pkgChart.Chart),
					Features:          pkg.Chart.Features,
					MultiSelect:       pkg.Chart.MultiSelect,
					Namespace:         XorY(pkg.Chart.Namespace, bundle.Spec.Namespace),
					Required:          required,
					Selected:          pkg.Chart.Required,
				},
			}
			for _, v := range pkg.Chart.Versions {
				card.Chart.Versions = append(card.Chart.Versions, v.VersionOption)
			}
			if len(card.Chart.Versions) == 1 {
				card.Chart.Versions[0].Selected = true
			}
			bv.Packages = append(bv.Packages, card)
		} else if pkg.Bundle != nil {
			view, err := toBundleOptionView(reg, pkg.Bundle, level+1)
			if err != nil {
				return nil, err
			}
			bv.Packages = append(bv.Packages, releasesapi.PackageCard{
				Bundle: view,
			})
		} else if pkg.OneOf != nil {
			bovs := make([]*releasesapi.BundleOptionView, 0, len(pkg.OneOf.Bundles))
			for _, bo := range pkg.OneOf.Bundles {
				view, err := toBundleOptionView(reg, bo, level+1)
				if err != nil {
					return nil, err
				}
				bovs = append(bovs, view)
			}
			bv.Packages = append(bv.Packages, releasesapi.PackageCard{
				OneOf: &releasesapi.OneOfBundleOptionView{
					Description: pkg.OneOf.Description,
					Bundles:     bovs,
				},
			})
		}
	}

	return &bv, nil
}

func CreateBundleViewForChart(reg repo.IRegistry, ref *releasesapi.ChartRepoRef) (*releasesapi.BundleView, error) {
	pkgChart, err := reg.GetChart(ref.URL, ref.Name, ref.Version)
	if err != nil {
		return nil, err
	}

	_, _, err = getBundle(pkgChart.Chart)
	if err == nil {
		return CreateBundleViewForBundle(reg, ref)
	} else if !kerr.IsNotFound(err) {
		return nil, err
	}

	return &releasesapi.BundleView{
		TypeMeta: metav1.TypeMeta{
			APIVersion: releasesapi.GroupVersion.String(),
			Kind:       "BundleView",
		},
		BundleOptionView: releasesapi.BundleOptionView{
			PackageMeta: releasesapi.PackageMeta{
				PackageDescriptor: GetPackageDescriptor(pkgChart.Chart),
				URL:               ref.URL,
				Name:              ref.Name,
				Version:           ref.Version,
			},
			DisplayName: flect.Titleize(flect.Humanize(ref.Name)),
			// Features:    nil,
			Packages: []releasesapi.PackageCard{
				{
					Chart: &releasesapi.ChartCard{
						ChartRef: releasesapi.ChartRef{
							URL:  ref.URL,
							Name: ref.Name,
						},
						PackageDescriptor: GetPackageDescriptor(pkgChart.Chart),
						Features:          []string{pkgChart.Metadata.Description},
						Namespace:         "default",
						Versions: []releasesapi.VersionOption{
							{
								Version:  ref.Version,
								Selected: true,
							},
						},
						MultiSelect: false,
						Required:    true,
						Selected:    true,
					},
				},
			},
		},
	}, nil
}
