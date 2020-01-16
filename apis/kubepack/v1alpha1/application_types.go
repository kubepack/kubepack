/*
Copyright The Kubepack Authors.

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
)

const (
	ResourceKindApplication = "Application"
	ResourceApplication     = "application"
	ResourceApplications    = "applications"
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

// Descriptor defines the Metadata and informations about the Application.
type Descriptor struct {
	// Type is the type of the application (e.g. WordPress, MySQL, Cassandra).
	Type string `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`

	// Version is an optional version indicator for the Application.
	Version string `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`

	// Description is a brief string description of the Application.
	Description string `json:"description,omitempty" protobuf:"bytes,3,opt,name=description"`

	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type.
	Icons []ImageSpec `json:"icons,omitempty" protobuf:"bytes,4,rep,name=icons"`

	// Maintainers is an optional list of maintainers of the application. The maintainers in this list maintain the
	// the source code, images, and package for the application.
	Maintainers []ContactData `json:"maintainers,omitempty" protobuf:"bytes,5,rep,name=maintainers"`

	// Owners is an optional list of the owners of the installed application. The owners of the application should be
	// contacted in the event of a orderned or unorderned disruption affecting the application.
	Owners []ContactData `json:"owners,omitempty" protobuf:"bytes,6,rep,name=owners"`

	// Keywords is an optional list of key words associated with the application (e.g. MySQL, RDBMS, database).
	Keywords []string `json:"keywords,omitempty" protobuf:"bytes,7,rep,name=keywords"`

	// Links are a list of descriptive URLs intended to be used to surface additional documentation, dashboards, etc.
	Links []Link `json:"links,omitempty" protobuf:"bytes,8,rep,name=links"`

	// Notes contain a human readable snippets intended as a quick start for the users of the Application.
	// CommonMark markdown syntax may be used for rich text representation.
	Notes string `json:"notes,omitempty" protobuf:"bytes,9,opt,name=notes"`
}

// ApplicationSpec defines the specification for an Application.
type ApplicationSpec struct {
	// ComponentGroupKinds is a list of Kinds for Application's components (e.g. Deployments, Pods, Services, CRDs). It
	// can be used in conjunction with the Application's Selector to list or watch the Applications components.
	ComponentGroupKinds []metav1.GroupKind `json:"componentKinds,omitempty" protobuf:"bytes,1,rep,name=componentKinds"`

	// Description regroups information and metadata about an application.
	Description Descriptor `json:"description,omitempty" protobuf:"bytes,2,opt,name=description"`

	// Selector is a label query over kinds that created by the application. It must match the component objects' labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,3,opt,name=selector"`

	// AddOwnerRef objects - flag to indicate if we need to add OwnerRefs to matching objects
	// Matching is done by using Selector to query all ComponentGroupKinds
	AddOwnerRef bool `json:"addOwnerRef,omitempty" protobuf:"varint,4,opt,name=addOwnerRef"`

	// Info contains human readable key,value pairs for the Application.
	Info []InfoItem `json:"info,omitempty" protobuf:"bytes,5,rep,name=info"`

	// AssemblyPhase represents the current phase of the application's assembly.
	// An empty value is equivalent to "Succeeded".
	AssemblyPhase ApplicationAssemblyPhase `json:"assemblyPhase,omitempty" protobuf:"bytes,6,opt,name=assemblyPhase,casttype=ApplicationAssemblyPhase"`
}

// ComponentList is a generic status holder for the top level resource
// +k8s:deepcopy-gen=true
type ComponentList struct {
	// Object status array for all matching objects
	Objects []ObjectStatus `json:"components,omitempty" protobuf:"bytes,1,rep,name=components"`
}

// ObjectStatus is a generic status holder for objects
// +k8s:deepcopy-gen=true
type ObjectStatus struct {
	// Link to object
	Link string `json:"link,omitempty" protobuf:"bytes,1,opt,name=link"`
	// Name of object
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
	// Kind of object
	Kind string `json:"kind,omitempty" protobuf:"bytes,3,opt,name=kind"`
	// Object group
	Group string `json:"group,omitempty" protobuf:"bytes,4,opt,name=group"`
	// Status. Values: InProgress, Ready, Unknown
	Status string `json:"status,omitempty" protobuf:"bytes,5,opt,name=status"`
}

// ConditionType encodes information on the condition
type ConditionType string

// Condition describes the state of an object at a certain point.
// +k8s:deepcopy-gen=true
type Condition struct {
	// Type of condition.
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	// Last time the condition was probed
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,5,opt,name=lastUpdateTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,6,opt,name=lastTransitionTime"`
}

// ApplicationStatus defines controller's the observed state of Application
type ApplicationStatus struct {
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
	ComponentList `json:",inline,omitempty" protobuf:"bytes,11,opt,name=componentList"`
}

// ImageSpec contains information about an image used as an icon.
type ImageSpec struct {
	// The source for image represented as either an absolute URL to the image or a Data URL containing
	// the image. Data URLs are defined in RFC 2397.
	Source string `json:"src" protobuf:"bytes,1,opt,name=src"`

	// (optional) The size of the image in pixels (e.g., 25x25).
	TotalSize string `json:"size,omitempty" protobuf:"bytes,2,opt,name=size"`

	// (optional) The mine type of the image (e.g., "image/png").
	Type string `json:"type,omitempty" protobuf:"bytes,3,opt,name=type"`
}

// ContactData contains information about an individual or organization.
type ContactData struct {
	// Name is the descriptive name.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Url could typically be a website address.
	URL string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`

	// Email is the email address.
	Email string `json:"email,omitempty" protobuf:"bytes,3,opt,name=email"`
}

// Link contains information about an URL to surface documentation, dashboards, etc.
type Link struct {
	// Description is human readable content explaining the purpose of the link.
	Description string `json:"description,omitempty" protobuf:"bytes,1,opt,name=description"`

	// Url typically points at a website address.
	URL string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
}

// InfoItem is a human readable key,value pair containing important information about how to access the Application.
type InfoItem struct {
	// Name is a human readable title for this piece of information.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Type of the value for this InfoItem.
	Type InfoItemType `json:"type,omitempty" protobuf:"bytes,2,opt,name=type,casttype=InfoItemType"`

	// Value is human readable content.
	Value string `json:"value,omitempty" protobuf:"bytes,3,opt,name=value"`

	// ValueFrom defines a reference to derive the value from another source.
	ValueFrom *InfoItemSource `json:"valueFrom,omitempty" protobuf:"bytes,4,opt,name=valueFrom"`
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
	Type InfoItemSourceType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=InfoItemSourceType"`

	// Selects a key of a Secret.
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty" protobuf:"bytes,2,opt,name=secretKeyRef"`

	// Selects a key of a ConfigMap.
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty" protobuf:"bytes,3,opt,name=configMapKeyRef"`

	// Select a Service.
	ServiceRef *ServiceSelector `json:"serviceRef,omitempty" protobuf:"bytes,4,opt,name=serviceRef"`

	// Select an Ingress.
	IngressRef *IngressSelector `json:"ingressRef,omitempty" protobuf:"bytes,5,opt,name=ingressRef"`
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
	core.ObjectReference `json:",inline" protobuf:"bytes,1,opt,name=objectReference"`
	// The key to select.
	Key string `json:"key,omitempty" protobuf:"bytes,2,opt,name=key"`
}

// SecretKeySelector selects a key from a Secret.
type SecretKeySelector struct {
	// The Secret to select from.
	core.ObjectReference `json:",inline" protobuf:"bytes,1,opt,name=objectReference"`
	// The key to select.
	Key string `json:"key,omitempty" protobuf:"bytes,2,opt,name=key"`
}

// ServiceSelector selects a Service.
type ServiceSelector struct {
	// The Service to select from.
	core.ObjectReference `json:",inline" protobuf:"bytes,1,opt,name=objectReference"`
	// The optional port to select.
	Port *int32 `json:"port,omitempty" protobuf:"varint,2,opt,name=port"`
	// The optional HTTP path.
	Path string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
}

// IngressSelector selects an Ingress.
type IngressSelector struct {
	// The Ingress to select from.
	core.ObjectReference `json:",inline" protobuf:"bytes,1,opt,name=objectReference"`
	// The optional host to select.
	Host string `json:"host,omitempty" protobuf:"bytes,2,opt,name=host"`
	// The optional HTTP path.
	Path string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
}

// ApplicationAssemblyPhase tracks the Application CRD phases: pending, succeded, failed
type ApplicationAssemblyPhase string

// Constants
const (
	// Used to indicate that not all of application's components
	// have been deployed yet.
	Pending ApplicationAssemblyPhase = "Pending"
	// Used to indicate that all of application's components
	// have already been deployed.
	Succeeded = "Succeeded"
	// Used to indicate that deployment of application's components
	// failed. Some components might be present, but deployment of
	// the remaining ones will not be re-attempted.
	Failed = "Failed"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=applications,singular=application,shortName=app,categories={kubepack,appscode}
// +kubebuilder:subresource:status

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ApplicationSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ApplicationStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Application `json:"items" protobuf:"bytes,2,rep,name=items"`
}
