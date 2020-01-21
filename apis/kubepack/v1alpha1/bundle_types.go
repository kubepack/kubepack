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
	"k8s.io/apimachinery/pkg/runtime"
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
	DisplayName       string       `json:"displayName" protobuf:"bytes,4,opt,name=displayName"`
	Features          []Feature    `json:"features,omitempty" protobuf:"bytes,5,rep,name=features"`
	Namespace         string       `json:"namespace,omitempty" protobuf:"bytes,6,opt,name=namespace"`
	Packages          []PackageRef `json:"packages" protobuf:"bytes,7,rep,name=packages"`
	Product           *ProductRef  `json:"product" protobuf:"bytes,8,opt,name=product"`
}

type PackageRef struct {
	Chart  *ChartOption       `json:"chart,omitempty" protobuf:"bytes,1,opt,name=chart"`
	Bundle *BundleOption      `json:"bundle,omitempty" protobuf:"bytes,2,opt,name=bundle"`
	OneOf  *OneOfBundleOption `json:"oneOf,omitempty" protobuf:"bytes,3,rep,name=oneOf"`
}

type OneOfBundleOption struct {
	Description string          `json:"description" protobuf:"bytes,1,opt,name=description"`
	Bundles     []*BundleOption `json:"bundles,omitempty" protobuf:"bytes,2,rep,name=bundles"`
}

type ChartRef struct {
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

type SelectionMode string

type ChartOption struct {
	ChartRef    `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`
	Features    []string        `json:"features,omitempty" protobuf:"bytes,2,rep,name=features"`
	Namespace   string          `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	Versions    []VersionDetail `json:"versions" protobuf:"bytes,4,rep,name=versions"`
	MultiSelect bool            `json:"multiSelect,omitempty" protobuf:"varint,5,opt,name=multiSelect"`
	Required    bool            `json:"required,omitempty" protobuf:"varint,6,opt,name=required"`
	Selected    bool            `json:"selected,omitempty" protobuf:"varint,7,opt,name=selected"`
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
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:validation:EmbeddedResource
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"parameters,omitempty" protobuf:"bytes,3,opt,name=parameters"`
}

type VersionDetail struct {
	VersionOption `json:",inline" protobuf:"bytes,1,opt,name=versionOption"`
	Resources     *ResourceDefinitions `json:"resources,omitempty" protobuf:"bytes,3,opt,name=resources"`
	WaitFors      []WaitFlags          `json:"waitFors,omitempty" protobuf:"bytes,4,rep,name=waitFors"`
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
