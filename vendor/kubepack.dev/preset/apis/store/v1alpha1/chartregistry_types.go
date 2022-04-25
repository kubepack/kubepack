/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	"kubepack.dev/preset/crds"
)

const (
	ResourceKindPreset      = "ChartRegistry"
	ResourceChartRegistry   = "chartregistry"
	ResourceChartRegistries = "chartregistries"
)

// ChartRegistrySpec defines the desired state of ChartRegistry
type ChartRegistrySpec struct {
	URL       string                 `json:"url"`
	SecretRef *kmapi.ObjectReference `json:"secretRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ChartRegistry is the Schema for the chartregistries API
type ChartRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChartRegistrySpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// ChartRegistryList contains a list of ChartRegistry
type ChartRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChartRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChartRegistry{}, &ChartRegistryList{})
}

func (_ ChartRegistry) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourceChartRegistries))
}
