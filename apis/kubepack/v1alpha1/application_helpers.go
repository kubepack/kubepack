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

package v1alpha1

import (
	"fmt"

	"kmodules.xyz/client-go/apiextensions"
	"sigs.k8s.io/application/api/app/v1beta1"
	"sigs.k8s.io/yaml"
)

func ConvertImageSpec(in []ImageSpec) []v1beta1.ImageSpec {
	out := make([]v1beta1.ImageSpec, len(in))
	for i := range in {
		out[i] = v1beta1.ImageSpec{
			Source: in[i].Source,
			Size:   in[i].TotalSize,
			Type:   in[i].Type,
		}
	}
	return out
}

func ConvertContactData(in []ContactData) []v1beta1.ContactData {
	out := make([]v1beta1.ContactData, len(in))
	for i := range in {
		out[i] = v1beta1.ContactData{
			Name:  in[i].Name,
			URL:   in[i].URL,
			Email: in[i].Email,
		}
	}
	return out
}

func ConvertLink(in []Link) []v1beta1.Link {
	out := make([]v1beta1.Link, len(in))
	for i := range in {
		out[i] = v1beta1.Link{
			Description: string(in[i].Description),
			URL:         in[i].URL,
		}
	}
	return out
}

func ApplicationCustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	var out apiextensions.CustomResourceDefinition

	v1file := fmt.Sprintf("%s_%s.yaml", v1beta1.GroupVersion.Group, "applications")
	if err := yaml.Unmarshal(v1beta1.MustAsset(v1file), &out.V1); err != nil {
		panic(err)
	}

	if out.V1 == nil && out.V1beta1 == nil {
		panic(fmt.Errorf("missing crd yamls for gvr: %s", v1beta1.GroupVersion.WithResource("applications")))
	}

	return &out
}
