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
	PackageDescriptor `json:",inline" protobuf:"bytes,3,opt,name=packageDescriptor"`
	Packages          []PackageRef `json:"packages" protobuf:"bytes,1,rep,name=packages"`
}

type PackageRef struct {
	Chart    *ChartOption    `json:"chart,omitempty" protobuf:"bytes,1,opt,name=chart"`
	Bundle   *BundleOption   `json:"bundle,omitempty" protobuf:"bytes,2,opt,name=bundle"`
	OneOf    []*BundleOption `json:"oneOf,omitempty" protobuf:"bytes,3,rep,name=oneOf"`
	Required bool            `json:"required,omitempty" protobuf:"varint,4,opt,name=required"`
}

type ChartRef struct {
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// One sentence description of feature provided by this addon
	// TODO: Move to a different struct instead of ref
	Features []string `json:"features,omitempty" protobuf:"bytes,3,rep,name=features"`
}

type SelectionMode string

type ChartOption struct {
	ChartRef    `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`
	Versions    []VersionOption `json:"versions" protobuf:"bytes,2,rep,name=versions"`
	MultiSelect bool            `json:"multiSelect,omitempty" protobuf:"varint,3,opt,name=multiSelect"`
}

type ChartVersionRef struct {
	ChartRef `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`
	Versions []string `json:"versions" protobuf:"bytes,2,rep,name=versions"`
}

type BundleRef struct {
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

type BundleOption struct {
	BundleRef `json:",inline" protobuf:"bytes,1,opt,name=bundleRef"`
	Version   string `json:"version" protobuf:"bytes,2,opt,name=version"`
}

type VersionOption struct {
	Version  string `json:"version" protobuf:"bytes,1,opt,name=version"`
	Selected bool   `json:"selected,omitempty" protobuf:"varint,2,opt,name=selected"`
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