/*
Copyright The Kmodules Authors.

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
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindResourceDescriptor = "ResourceDescriptor"
	ResourceResourceDescriptor     = "resourcedescriptor"
	ResourceResourceDescriptors    = "resourcedescriptors"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=resourcedescriptors,singular=resourcedescriptor,shortName=rd
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ResourceDescriptor struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceDescriptorSpec `json:"spec,omitempty"`
}

type ResourceDescriptorSpec struct {
	Resource    ResourceID                   `json:"resource"`
	Columns     []ResourceColumnDefinition   `json:"columns,omitempty"`
	SubTables   []ResourceSubTableDefinition `json:"subTables,omitempty"`
	Connections []ResourceConnection         `json:"connections,omitempty"`
	KeyTargets  []metav1.TypeMeta            `json:"keyTargets,omitempty"`

	// validation describes the schema used for validation and pruning of the custom resource.
	// If present, this validation schema is used to validate all versions.
	// Top-level and per-version schemas are mutually exclusive.
	// +optional
	Validation *crdv1beta1.CustomResourceValidation `json:"validation,omitempty"`

	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type.
	Icons []ImageSpec `json:"icons,omitempty"`

	// Maintainers is an optional list of maintainers of the application. The maintainers in this list maintain the
	// the source code, images, and package for the application.
	Maintainers []ContactData `json:"maintainers,omitempty"`

	// Links are a list of descriptive URLs intended to be used to surface additional documentation, dashboards, etc.
	Links []Link `json:"links,omitempty"`

	Installer *DeploymentParameters `json:"installer,omitempty"`
}

type DeploymentParameters struct {
	ProductID string    `json:"productID,omitempty"`
	PlanID    string    `json:"planID,omitempty"`
	Version   string    `json:"version,omitempty"`
	Chart     *ChartRef `json:"chart,omitempty"`
}

type ChartRef struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// ResourceID identifies a resource
type ResourceID struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	// Name is the plural name of the resource to serve.  It must match the name of the CustomResourceDefinition-registration
	// too: plural.group and it must be all lowercase.
	Name string `json:"name"`
	// Kind is the serialized kind of the resource.  It is normally CamelCase and singular.
	Kind  string        `json:"kind"`
	Scope ResourceScope `json:"scope"`
}

// ResourceScope is an enum defining the different scopes available to a custom resource
type ResourceScope string

const (
	ClusterScoped   ResourceScope = "Cluster"
	NamespaceScoped ResourceScope = "Namespaced"
)

type ConnectionType string

const (
	MatchSelector ConnectionType = "MatchSelector"
	MatchName     ConnectionType = "MatchName"
	MatchRef      ConnectionType = "MatchRef"
	OwnedBy       ConnectionType = "OwnedBy"
)

type ResourceConnection struct {
	Target                 metav1.TypeMeta `json:"target"`
	ResourceConnectionSpec `json:",inline,omitempty"`
}

type ResourceConnectionSpec struct {
	Type          ConnectionType `json:"type"`
	NamespacePath string         `json:"namespacePath,omitempty"`

	// default: metadata.labels
	// +optional
	TargetLabelPath string                `json:"targetLabelPath,omitempty"`
	SelectorPath    string                `json:"selectorPath,omitempty"`
	Selector        *metav1.LabelSelector `json:"selector,omitempty"`

	NameTemplate string `json:"nameTemplate,omitempty"`

	// References are a jsonpath that returns a CSV formatted references to target resources
	//
	// If each row has a single column, it is target name. Target resource is non-namespaced or
	// uses the same namespace as the source resource. Example:
	// n1
	// n2
	//
	// If each row has two columns, it is target [name,namespace]. Example:
	// n1,ns1
	// n2,ns2
	//
	// If each row has three columns, it is target [name,namespace,kind]. Example:
	// n1,ns1,k1
	// n2,ns2,k2
	//
	// If each row has four columns, it is target [name,namespace,kind,apiGroup]. Example:
	// n1,ns1,k1,apiGroup1
	// n2,ns2,k2,apiGroup2
	References []string `json:"references,omitempty"`

	Level OwnershipLevel `json:"level,omitempty"`
}

type OwnershipLevel string

const (
	Reference  OwnershipLevel = ""
	Owner      OwnershipLevel = "Owner"
	Controller OwnershipLevel = "Controller"
)

type Priority int32

const (
	Field Priority = 1 << iota
	List
)

// ResourceColumnDefinition specifies a column for server side printing.
type ResourceColumnDefinition struct {
	// name is a human readable name for the column.
	Name string `json:"name"`
	// type is an OpenAPI type definition for this column.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	Type string `json:"type"`
	// format is an optional OpenAPI type definition for this column. The 'name' format is applied
	// to the primary identifier column to assist in clients identifying column is the resource name.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	// +optional
	Format string `json:"format,omitempty"`
	// description is a human readable description of this column.
	// +optional
	Description string `json:"description,omitempty"`
	// priority is an integer defining the relative importance of this column compared to others. Lower
	// numbers are considered higher priority. Columns that may be omitted in limited space scenarios
	// should be given a higher priority.
	Priority int32 `json:"priority"`
	// JSONPath is a simple JSON path, i.e. with array notation.
	JSONPath string `json:"jsonPath"`
}

type ResourceSubTableDefinition struct {
	Name      string                     `json:"name"`
	FieldPath string                     `json:"fieldPath,omitempty"`
	Columns   []ResourceColumnDefinition `json:"columns,omitempty"`
}

// ImageSpec contains information about an image used as an icon.
type ImageSpec struct {
	// The source for image represented as either an absolute URL to the image or a Data URL containing
	// the image. Data URLs are defined in RFC 2397.
	Source string `json:"src"`

	// (optional) The size of the image in pixels (e.g., 25x25).
	Size string `json:"size,omitempty"`

	// (optional) The mine type of the image (e.g., "image/png").
	Type string `json:"type,omitempty"`
}

// ContactData contains information about an individual or organization.
type ContactData struct {
	// Name is the descriptive name.
	Name string `json:"name,omitempty"`

	// Url could typically be a website address.
	URL string `json:"url,omitempty"`

	// Email is the email address.
	Email string `json:"email,omitempty"`
}

// Link contains information about an URL to surface documentation, dashboards, etc.
type Link struct {
	// Description is human readable content explaining the purpose of the link.
	Description string `json:"description,omitempty"`

	// Url typically points at a website address.
	URL string `json:"url,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ResourceDescriptorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceDescriptor `json:"items,omitempty"`
}
