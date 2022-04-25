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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kmodules.xyz/client-go/apiextensions"
	"kubepack.dev/preset/crds"
)

const (
	ResourceKindClusterChartPreset = "ClusterChartPreset"
	ResourceClusterChartPreset     = "clusterchartpreset"
	ResourceClusterChartPresets    = "clusterchartpresets"
)

// ClusterChartPresetSpec defines the desired state of ClusterChartPreset
type ClusterChartPresetSpec struct {
	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// +optional
	UsePresets []core.TypedLocalObjectReference `json:"usePresets,omitempty"`

	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Values *runtime.RawExtension `json:"values,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ClusterChartPreset is the Schema for the clusterchartpresets API
type ClusterChartPreset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterChartPresetSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterChartPresetList contains a list of ClusterChartPreset
type ClusterChartPresetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterChartPreset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterChartPreset{}, &ClusterChartPresetList{})
}

func (_ ClusterChartPreset) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourceClusterChartPresets))
}
