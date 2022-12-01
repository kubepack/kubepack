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
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metadata/apis/shared"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindResourceEditor = "ResourceEditor"
	ResourceResourceEditor     = "resourceeditor"
	ResourceResourceEditors    = "resourceeditors"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=resourceeditors,singular=resourceeditor,scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ResourceEditor struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceEditorSpec `json:"spec,omitempty"`
}

type ResourceEditorSpec struct {
	Resource kmapi.ResourceID `json:"resource"`
	UI       *UIParameters    `json:"ui,omitempty"`
	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type.
	Icons []shared.ImageSpec `json:"icons,omitempty"`
	// Kind == VendorChartPreset | ClusterChartPreset
	Variants  []VariantRef                 `json:"variants,omitempty"`
	Installer *shared.DeploymentParameters `json:"installer,omitempty"`
}

type UIParameters struct {
	Options *shared.ChartRepoRef `json:"options,omitempty"`
	Editor  *shared.ChartRepoRef `json:"editor,omitempty"`
	// app.kubernetes.io/instance label must be updated at these paths when refilling metadata
	// +optional
	InstanceLabelPaths []string `json:"instanceLabelPaths,omitempty"`
	// +optional
	Actions []*ActionTemplateGroup `json:"actions,omitempty"`
}

type ActionTemplateGroup struct {
	shared.ActionInfo `json:",inline,omitempty"`
	Items             []ActionTemplate `json:"items"`
}

type ActionTemplate struct {
	shared.ActionInfo `json:",inline,omitempty"`
	// +optional
	Icons            []shared.ImageSpec   `json:"icons,omitempty"`
	OperationID      string               `json:"operationId"`
	Flow             string               `json:"flow"`
	DisabledTemplate string               `json:"disabledTemplate,omitempty"`
	Editor           *shared.ChartRepoRef `json:"editor,omitempty"`
}

type VariantRef struct {
	core.TypedLocalObjectReference `json:",inline"`

	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type.
	Icons []shared.ImageSpec `json:"icons,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ResourceEditorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceEditor `json:"items,omitempty"`
}
