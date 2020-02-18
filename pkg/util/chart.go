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
	"net/http"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/lib-helm/repo"

	"github.com/gabriel-vasile/mimetype"
	"helm.sh/helm/v3/pkg/chart"
)

func GetPackageDescriptor(pkgChart *chart.Chart) v1alpha1.PackageDescriptor {
	var out v1alpha1.PackageDescriptor

	out.Description = pkgChart.Metadata.Description
	if pkgChart.Metadata.Icon != "" {
		var imgType string
		if resp, err := http.Get(pkgChart.Metadata.Icon); err == nil {
			if mime, err := mimetype.DetectReader(resp.Body); err == nil {
				imgType = mime.String()
			}
			_ = resp.Body.Close()
		}
		out.Icons = []v1alpha1.ImageSpec{
			{
				Source: pkgChart.Metadata.Icon,
				// TotalSize: "",
				Type: imgType,
			},
		}
	}
	for _, maintainer := range pkgChart.Metadata.Maintainers {
		out.Maintainers = append(out.Maintainers, v1alpha1.ContactData{
			Name:  maintainer.Name,
			URL:   maintainer.URL,
			Email: maintainer.Email,
		})
	}
	out.Keywords = pkgChart.Metadata.Keywords

	if pkgChart.Metadata.Home != "" {
		out.Links = []v1alpha1.Link{
			{
				Description: v1alpha1.LinkWebsite,
				URL:         pkgChart.Metadata.Home,
			},
		}
	}

	return out
}

var reg = repo.NewDiskCacheRegistry()

func GetChart(repoURL, chartName, chartVersion string) (*repo.ChartExtended, error) {
	return reg.GetChart(repoURL, chartName, chartVersion)
}
