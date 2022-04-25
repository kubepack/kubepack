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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindResourceDashboard = "ResourceDashboard"
	ResourceResourceDashboard     = "resourcedashboard"
	ResourceResourceDashboards    = "resourcedashboards"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=resourcedashboards,singular=resourcedashboard
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ResourceDashboard struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceDashboardSpec `json:"spec,omitempty"`
}

// +kubebuilder:validation:Enum=Grafana
type DashboardProvider string

const (
	DashboardProviderGrafana DashboardProvider = "Grafana"
)

type ResourceDashboardSpec struct {
	Resource   kmapi.ResourceID  `json:"resource"`
	Provider   DashboardProvider `json:"provider,omitempty"`
	Dashboards []Dashboard       `json:"dashboards"`
}

// +kubebuilder:validation:Enum=Source;Target
type DashboardVarType string

const (
	DashboardVarTypeSource DashboardVarType = "Source"
	DashboardVarTypeTarget DashboardVarType = "Target"
)

type DashboardVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	// +optional
	// +kubebuilder:default:=Source
	Type DashboardVarType `json:"type,omitempty"`
}

type Dashboard struct {
	// +optional
	Title string `json:"title,omitempty"`
	// +optional
	Vars []DashboardVar `json:"vars,omitempty"`
	// +optional
	Panels []string `json:"panels,omitempty"`
	// +optional
	If *If `json:"if,omitempty"`
}

type If struct {
	Condition string           `json:"condition,omitempty"`
	Connected *ResourceLocator `json:"connected,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ResourceDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceDashboard `json:"items,omitempty"`
}
