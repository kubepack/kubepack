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
	"strings"

	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/resource-metadata/crds"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r ResourceID) GroupVersion() schema.GroupVersion {
	return schema.GroupVersion{Group: r.Group, Version: r.Version}
}

func (r ResourceID) GroupResource() schema.GroupResource {
	return schema.GroupResource{Group: r.Group, Resource: r.Name}
}

func (r ResourceID) TypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{APIVersion: r.GroupVersion().String(), Kind: r.Kind}
}

func (r ResourceID) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: r.Group, Version: r.Version, Resource: r.Name}
}

func (r ResourceID) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: r.Group, Version: r.Version, Kind: r.Kind}
}

func (v ResourceDescriptor) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceResourceDescriptors))
}

func (v ResourceDescriptor) IsValid() error {
	return nil
}

func IsOfficialType(group string) bool {
	switch {
	case group == "":
		return true
	case !strings.ContainsRune(group, '.'):
		return true
	case group == "k8s.io" || strings.HasSuffix(group, ".k8s.io"):
		return true
	case group == "kubernetes.io" || strings.HasSuffix(group, ".kubernetes.io"):
		return true
	case group == "x-k8s.io" || strings.HasSuffix(group, ".x-k8s.io"):
		return true
	default:
		return false
	}
}
