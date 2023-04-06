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
	"kmodules.xyz/resource-metadata/apis/shared"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindFeatureSet = "FeatureSet"
	ResourceFeatureSet     = "featureset"
	ResourceFeatureSets    = "featuresets"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=featuresets,singular=featureset,scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Required",type="boolean",JSONPath=".spec.required"
// +kubebuilder:printcolumn:name="Enabled",type="boolean",JSONPath=".status.enabled"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type FeatureSet struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FeatureSetSpec   `json:"spec,omitempty"`
	Status            FeatureSetStatus `json:"status,omitempty"`
}

type FeatureSetSpec struct {
	// Title specify the title of this feature set.
	Title string `json:"title"`
	// Description specifies a short description of the services this feature set provides.
	Description string `json:"description"`
	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type. These icons will be used in UI.
	Icons []shared.ImageSpec `json:"icons,omitempty"`
	// Required specify whether this feature set is mandatory or not for using the UI.
	// +optional
	Required bool `json:"required,omitempty"`
	// RequiredFeatures specifies list of features that are necessary to consider this feature set as ready.
	// +optional
	RequiredFeatures []string `json:"requiredFeatures,omitempty"`
	// Chart specifies the chart that contains the respective resources for component features and the UI wizard.
	Chart shared.ExpandedChartRepoRef `json:"chart"`
}

type FeatureSetStatus struct {
	// Enabled specifies whether this feature set is enabled or not.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// Ready specifies whether this feature set is ready not.
	// +optional
	Ready *bool `json:"ready,omitempty"`
	// Features specifies the status of the component features that belong to this feature set.
	// +optional
	Features []ComponentStatus `json:"features,omitempty"`
	// Note specifies the respective reason if the feature set is considered as disabled.
	// +optional
	Note string `json:"note,omitempty"`
}

type ComponentStatus struct {
	// Name specify the name of the component feature.
	Name string `json:"name"`
	// Enabled specifies whether the component feature has been enabled or not.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// Ready specifies whether the component feature is ready or not.
	// +optional
	Ready *bool `json:"ready,omitempty"`
	// Managed specifies whether the component is managed by platform or not.
	// +optional
	Managed *bool `json:"managed,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type FeatureSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureSet `json:"items,omitempty"`
}
