// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
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

// Constants for condition
const (
	// Ready => controller considers this resource Ready
	Ready = "Ready"
	// Qualified => functionally tested
	Qualified = "Qualified"
	// Settled => observed generation == generation + settled means controller is done acting functionally tested
	Settled = "Settled"
	// Cleanup => it is set to track finalizer failures
	Cleanup = "Cleanup"
	// Error => last recorded error
	Error = "Error"

	ReasonInit = "Init"
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
	// ComponentGroupKinds is a list of Kinds for AppRelease's components (e.g. Deployments, Pods, Services, CRDs). It
	// can be used in conjunction with the AppRelease's Selector to list or watch the AppReleases components.
	ComponentGroupKinds []metav1.GroupKind `json:"componentKinds,omitempty"`

	// Descriptor regroups information and metadata about an appRelease.
	Descriptor Descriptor `json:"descriptor,omitempty"`

	// Selector is a label query over kinds that created by the appRelease. It must match the component objects' labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// AddOwnerRef objects - flag to indicate if we need to add OwnerRefs to matching objects
	// Matching is done by using Selector to query all ComponentGroupKinds
	AddOwnerRef bool `json:"addOwnerRef,omitempty"`

	// Info contains human readable key,value pairs for the AppRelease.
	// +patchStrategy=merge
	// +patchMergeKey=name
	Info []InfoItem `json:"info,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// AssemblyPhase represents the current phase of the appRelease's assembly.
	// An empty value is equivalent to "Succeeded".
	AssemblyPhase AppReleaseAssemblyPhase `json:"assemblyPhase,omitempty"`
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

// ConditionType encodes information on the condition
type ConditionType string

// Condition describes the state of an object at a certain point.
type Condition struct {
	// Type of condition.
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=StatefulSetConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
	// Last time the condition was probed
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
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
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`
	// Resources embeds a list of object statuses
	// +optional
	ComponentList `json:",inline,omitempty"`
	// ComponentsReady: status of the components in the format ready/total
	// +optional
	ComponentsReady string `json:"componentsReady,omitempty"`
}

// InfoItem is a human readable key,value pair containing important information about how to access the AppRelease.
type InfoItem struct {
	// Name is a human readable title for this piece of information.
	Name string `json:"name,omitempty"`

	// Type of the value for this InfoItem.
	Type InfoItemType `json:"type,omitempty"`

	// Value is human readable content.
	Value string `json:"value,omitempty"`

	// ValueFrom defines a reference to derive the value from another source.
	ValueFrom *InfoItemSource `json:"valueFrom,omitempty"`
}

// InfoItemType is a string that describes the value of InfoItem
type InfoItemType string

const (
	// ValueInfoItemType const string for value type
	ValueInfoItemType InfoItemType = "Value"
	// ReferenceInfoItemType const string for ref type
	ReferenceInfoItemType InfoItemType = "Reference"
)

// InfoItemSource represents a source for the value of an InfoItem.
type InfoItemSource struct {
	// Type of source.
	Type InfoItemSourceType `json:"type,omitempty"`

	// Selects a key of a Secret.
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`

	// Selects a key of a ConfigMap.
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`

	// Select a Service.
	ServiceRef *ServiceSelector `json:"serviceRef,omitempty"`

	// Select an Ingress.
	IngressRef *IngressSelector `json:"ingressRef,omitempty"`
}

// InfoItemSourceType is a string
type InfoItemSourceType string

// Constants for info type
const (
	SecretKeyRefInfoItemSourceType    InfoItemSourceType = "SecretKeyRef"
	ConfigMapKeyRefInfoItemSourceType InfoItemSourceType = "ConfigMapKeyRef"
	ServiceRefInfoItemSourceType      InfoItemSourceType = "ServiceRef"
	IngressRefInfoItemSourceType      InfoItemSourceType = "IngressRef"
)

// ConfigMapKeySelector selects a key from a ConfigMap.
type ConfigMapKeySelector struct {
	// The ConfigMap to select from.
	corev1.ObjectReference `json:",inline"`
	// The key to select.
	Key string `json:"key,omitempty"`
}

// SecretKeySelector selects a key from a Secret.
type SecretKeySelector struct {
	// The Secret to select from.
	corev1.ObjectReference `json:",inline"`
	// The key to select.
	Key string `json:"key,omitempty"`
}

// ServiceSelector selects a Service.
type ServiceSelector struct {
	// The Service to select from.
	corev1.ObjectReference `json:",inline"`
	// The optional port to select.
	Port *int32 `json:"port,omitempty"`
	// The optional HTTP path.
	Path string `json:"path,omitempty"`
	// Protocol for the service
	Protocol string `json:"protocol,omitempty"`
}

// IngressSelector selects an Ingress.
type IngressSelector struct {
	// The Ingress to select from.
	corev1.ObjectReference `json:",inline"`
	// The optional host to select.
	Host string `json:"host,omitempty"`
	// The optional HTTP path.
	Path string `json:"path,omitempty"`
	// Protocol for the ingress
	Protocol string `json:"protocol,omitempty"`
}

// AppReleaseAssemblyPhase tracks the AppRelease CRD phases: pending, succeeded, failed
type AppReleaseAssemblyPhase string

// Constants
const (
	// Used to indicate that not all of appRelease's components
	// have been deployed yet.
	Pending AppReleaseAssemblyPhase = "Pending"
	// Used to indicate that all of appRelease's components
	// have already been deployed.
	Succeeded = "Succeeded"
	// Used to indicate that deployment of appRelease's components
	// failed. Some components might be present, but deployment of
	// the remaining ones will not be re-attempted.
	Failed = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,description="The type of the appRelease",JSONPath=`.spec.descriptor.type`,priority=0
// +kubebuilder:printcolumn:name="Version",type=string,description="The creation date",JSONPath=`.spec.descriptor.version`,priority=0
// +kubebuilder:printcolumn:name="Owner",type=boolean,description="The appRelease object owns the matched resources",JSONPath=`.spec.addOwnerRef`,priority=0
// +kubebuilder:printcolumn:name="Ready",type=string,description="Numbers of components ready",JSONPath=`.status.componentsReady`,priority=0
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
