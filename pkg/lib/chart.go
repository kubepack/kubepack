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
	"bytes"
	"net/http"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	kmapi "kmodules.xyz/client-go/api/v1"
	yamllib "sigs.k8s.io/yaml"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
	"x-helm.dev/apimachinery/apis/shared"
)

func GetPackageDescriptor(pkgChart *chart.Chart) releasesapi.PackageDescriptor {
	var out releasesapi.PackageDescriptor

	out.Description = pkgChart.Metadata.Description
	if pkgChart.Metadata.Icon != "" {
		var imgType string
		if resp, err := http.Get(pkgChart.Metadata.Icon); err == nil {
			if mime, err := mimetype.DetectReader(resp.Body); err == nil {
				imgType = mime.String()
			}
			_ = resp.Body.Close()
		}
		out.Icons = []shared.ImageSpec{
			{
				Source: pkgChart.Metadata.Icon,
				// TotalSize: "",
				Type: imgType,
			},
		}
	}
	for _, maintainer := range pkgChart.Metadata.Maintainers {
		out.Maintainers = append(out.Maintainers, shared.ContactData{
			Name:  maintainer.Name,
			URL:   maintainer.URL,
			Email: maintainer.Email,
		})
	}
	out.Keywords = pkgChart.Metadata.Keywords

	if pkgChart.Metadata.Home != "" {
		out.Links = []shared.Link{
			{
				Description: string(shared.LinkWebsite),
				URL:         pkgChart.Metadata.Home,
			},
		}
	}

	return out
}

func CreatePackageView(srcRef kmapi.TypedObjectReference, chrt *chart.Chart) (*releasesapi.PackageView, error) {
	p := releasesapi.PackageView{
		TypeMeta: metav1.TypeMeta{
			APIVersion: releasesapi.GroupVersion.String(),
			Kind:       "PackageView",
		},
		PackageMeta: releasesapi.PackageMeta{
			ChartSourceRef: releasesapi.ChartSourceRef{
				Name:      chrt.Name(),
				Version:   chrt.Metadata.Version,
				SourceRef: srcRef,
			},
			PackageDescriptor: GetPackageDescriptor(chrt),
		},
	}

	for _, f := range chrt.Raw {
		if f.Name == chartutil.ValuesfileName || (strings.HasPrefix(f.Name, "values-") && strings.HasSuffix(f.Name, ".yaml")) {
			var values map[string]interface{}
			err := yamllib.Unmarshal(f.Data, &values)
			if err != nil {
				return nil, err
			}

			p.ValuesFiles = append(p.ValuesFiles, releasesapi.ValuesFile{
				Filename: f.Name,
				Values: &runtime.RawExtension{
					Object: &unstructured.Unstructured{Object: values},
				},
			})
		}
		if f.Name == "values.openapiv3_schema.json" || f.Name == "values.openapiv3_schema.yaml" || f.Name == "values.openapiv3_schema.yml" {
			var schema crdv1.JSONSchemaProps
			reader := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(f.Data), 2048)
			err := reader.Decode(&schema)
			if err != nil {
				return nil, err
			}
			p.OpenAPIV3Schema = &schema
		}
	}
	//if b.Schema == nil && len(pkgChart.Schema) > 0 {
	//	// TODO convert json schema to openapi schema v3
	//}
	return &p, nil
}
