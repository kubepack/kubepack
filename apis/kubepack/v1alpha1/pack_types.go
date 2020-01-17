/*
Copyright The Kubepack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by packlicable law or agreed to in writing, software
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
	ResourceKindPack = "Package"
	ResourcePack     = "package"
	ResourcePacks    = "packages"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=packages,singular=package,scope=Cluster,shortName=pkg,categories={kubepack,appscode}
// +kubebuilder:subresource:status
type Pack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              PackSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            PackStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type PackSpec struct {
	Chart     *ChartRef            `json:"chart" protobuf:"bytes,1,opt,name=chart"`
	Resources *ResourceDefinitions `json:"resources,omitempty" protobuf:"bytes,2,opt,name=resources"`
	WaitFors  []WaitOptions        `json:"waitFors,omitempty" protobuf:"bytes,3,rep,name=waitFors"`
}

type ResourceDefinitions struct {
	Owned    []ResourceID `json:"owned" protobuf:"bytes,1,rep,name=owned"`
	Required []ResourceID `json:"required" protobuf:"bytes,2,rep,name=required"`
}

// wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available]

type WaitOptions struct {
	Resource     GroupVersionResource `json:"resource" protobuf:"bytes,1,opt,name=resource"`
	Name         string               `json:"name" protobuf:"bytes,2,opt,name=name"`
	Labels       metav1.LabelSelector `json:"labels" protobuf:"bytes,3,opt,name=labels"`
	All          bool                 `json:"all" protobuf:"varint,4,opt,name=all"`
	Timeout      metav1.Duration      `json:"timeout" protobuf:"bytes,5,opt,name=timeout"`
	ForCondition string               `json:"for" protobuf:"bytes,6,opt,name=for"`
}

type GroupVersionResource struct {
	Group    string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version  string `json:"version" protobuf:"bytes,2,opt,name=version"`
	Resource string `json:"resource" protobuf:"bytes,3,opt,name=resource"`
}

type PackStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type PackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Pack `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
