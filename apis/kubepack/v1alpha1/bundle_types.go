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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindBundle = "Bundle"
	ResourceBundle     = "bundle"
	ResourceBundles    = "bundles"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=bundles,singular=bundle,scope=Cluster,categories={kubepack,appscode}
// +kubebuilder:subresource:status
type Bundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              BundleSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            BundleStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type BundleSpec struct {
	Packages []PacakeRef `json:"packages" protobuf:"bytes,1,rep,name=packages"`
	Addons   []Addon     `json:"addons" protobuf:"bytes,2,rep,name=addons"`
}

type PacakeRef struct {
	Chart    ChartOptions `json:"chart" protobuf:"bytes,1,opt,name=chart"`
	Required bool         `json:"required" protobuf:"varint,2,opt,name=required"`
}

type ChartRef struct {
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// One sentence description of feature provided by this addon
	// TODO: Move to a different struct instead of ref
	Feature string `json:"feature" protobuf:"bytes,3,opt,name=feature"`
}

type ChartOptions struct {
	ChartRef `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`
	Versions []VersionOption `json:"versions" protobuf:"bytes,2,rep,name=versions"`
}

type ChartVersionRef struct {
	ChartRef `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`
	Versions []string `json:"versions" protobuf:"bytes,2,rep,name=versions"`
}

type BundleRef struct {
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// One sentence description of feature provided by this addon
	// TODO: Move to a different struct instead of ref
	Feature string `json:"feature" protobuf:"bytes,3,opt,name=feature"`
}

type BundleVersionRef struct {
	BundleRef `json:",inline" protobuf:"bytes,1,opt,name=bundleRef"`
	Versions  []string `json:"versions" protobuf:"bytes,2,rep,name=versions"`
}

type Addon struct {
	Feature string         `json:"feature" protobuf:"bytes,1,opt,name=feature"`
	Bundle  *BundleOption  `json:"bundle" protobuf:"bytes,2,opt,name=bundle"`
	OneOf   []BundleOption `json:"oneOf" protobuf:"bytes,3,rep,name=oneOf"`
}

type BundleOption struct {
	BundleRef `json:",inline" protobuf:"bytes,1,opt,name=bundleRef"`
	Versions  []VersionOption `json:"versions" protobuf:"bytes,2,rep,name=versions"`
}

type VersionOption struct {
	Version         string `json:"version" protobuf:"bytes,1,opt,name=version"`
	DefaultSelected bool   `json:"defaultSelected" protobuf:"varint,2,opt,name=defaultSelected"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type BundleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Bundle `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

type BundleStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
}
