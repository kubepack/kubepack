// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/client-go/apiextensions"
	"x-helm.dev/apimachinery/apis/shared"
	"x-helm.dev/apimachinery/crds"
)

const (
	ResourceKindAppRelease = "AppRelease"
	ResourceAppRelease     = "apprelease"
	ResourceAppReleases    = "appreleases"
)

// Descriptor defines the Metadata and informations about the AppRelease.
type Descriptor struct {
	// Type is the type of the appRelease (e.g. WordPress, MySQL, Cassandra).
	Type string `json:"type,omitempty"`

	// Version is an optional version indicator for the AppRelease.
	Version string `json:"version,omitempty"`

	// Description is a brief string description of the AppRelease.
	Description string `json:"description,omitempty"`

	// Icons is an optional list of icons for an appRelease. Icon information includes the source, size,
	// and mime type.
	Icons []shared.ImageSpec `json:"icons,omitempty"`

	// Maintainers is an optional list of maintainers of the appRelease. The maintainers in this list maintain the
	// the source code, images, and package for the appRelease.
	Maintainers []shared.ContactData `json:"maintainers,omitempty"`

	// Owners is an optional list of the owners of the installed appRelease. The owners of the appRelease should be
	// contacted in the event of a planned or unplanned disruption affecting the appRelease.
	Owners []shared.ContactData `json:"owners,omitempty"`

	// Keywords is an optional list of key words associated with the appRelease (e.g. MySQL, RDBMS, database).
	Keywords []string `json:"keywords,omitempty"`

	// Links are a list of descriptive URLs intended to be used to surface additional documentation, dashboards, etc.
	Links []shared.Link `json:"links,omitempty"`

	// Notes contain a human readable snippets intended as a quick start for the users of the AppRelease.
	// CommonMark markdown syntax may be used for rich text representation.
	Notes string `json:"notes,omitempty"`
}

// AppReleaseSpec defines the specification for an AppRelease.
type AppReleaseSpec struct {
	// Descriptor regroups information and metadata about an appRelease.
	Descriptor Descriptor `json:"descriptor,omitempty"`

	// Release regroups information and metadata about a Helm release.
	Release ReleaseInfo `json:"release,omitempty"`

	// Components is a list of Kinds for AppRelease's components (e.g. Deployments, Pods, Services, CRDs). It
	// can be used in conjunction with the AppRelease's Selector to list or watch the AppReleases components.
	Components []metav1.GroupVersionKind `json:"components,omitempty"`

	// Selector is a label query over kinds that created by the appRelease. It must match the component objects' labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	Editor *metav1.GroupVersionResource `json:"editor,omitempty"`

	// +optional
	ResourceKeys []string `json:"resourceKeys,omitempty"`

	// +optional
	FormKeys []string `json:"formKeys,omitempty"`
}

type ReleaseInfo struct {
	Name          string                `json:"name"`
	Version       string                `json:"version,omitempty"`
	Status        string                `json:"status,omitempty"`
	FirstDeployed *metav1.Time          `json:"firstDeployed,omitempty"`
	LastDeployed  *metav1.Time          `json:"lastDeployed,omitempty"`
	ModifiedAt    *metav1.Time          `json:"modified-at,omitempty"`
	Form          *runtime.RawExtension `json:"form,omitempty"`
}

// ComponentList is a generic status holder for the top level resource
type ComponentList struct {
	// Object status array for all matching objects
	Objects []ObjectStatus `json:"components,omitempty"`
}

// ObjectStatus is a generic status holder for objects
type ObjectStatus struct {
	// Link to object
	Link string `json:"link,omitempty"`
	// Name of object
	Name string `json:"name,omitempty"`
	// Kind of object
	Kind string `json:"kind,omitempty"`
	// Object group
	Group string `json:"group,omitempty"`
	// Status. Values: InProgress, Ready, Unknown
	Status string `json:"status,omitempty"`
}

// AppReleaseStatus defines controller's the observed state of AppRelease
type AppReleaseStatus struct {
	// ObservedGeneration is the most recent generation observed. It corresponds to the
	// Object's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
	// Conditions represents the latest state of the object
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`
	// Resources embeds a list of object statuses
	// +optional
	ComponentList `json:",inline,omitempty"`
	// ComponentsReady: status of the components in the format ready/total
	// +optional
	ComponentsReady string `json:"componentsReady,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,description="The type of the appRelease",JSONPath=`.spec.descriptor.type`,priority=0
// +kubebuilder:printcolumn:name="Version",type=string,description="The creation date",JSONPath=`.spec.descriptor.version`,priority=0
// +kubebuilder:printcolumn:name="Age",type=date,description="The creation date",JSONPath=`.metadata.creationTimestamp`,priority=0

// AppRelease is the Schema for the appReleases API
type AppRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppReleaseSpec   `json:"spec,omitempty"`
	Status AppReleaseStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AppReleaseList contains a list of AppRelease
type AppReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppRelease{}, &AppReleaseList{})
}

func (_ AppRelease) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourceAppReleases))
}

// StripVersion the version part of gv
func StripVersion(gv string) string {
	if gv == "" {
		return gv
	}

	re := regexp.MustCompile(`^[vV][0-9].*`)
	// If it begins with only version, (group is nil), return empty string which maps to core group
	if re.MatchString(gv) {
		return ""
	}

	return strings.Split(gv, "/")[0]
}
