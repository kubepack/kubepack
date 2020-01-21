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

// ChartRepoRef references to a single version of a Chart
type ChartRepoRef struct {
	Name    string `json:"name" protobuf:"bytes,1,opt,name=name"`
	URL     string `json:"url" protobuf:"bytes,2,opt,name=url"`
	Version string `json:"version" protobuf:"bytes,3,opt,name=version"`
}

// ResourceID identifies a resource
type ResourceID struct {
	Group   string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// Name is the plural name of the resource to serve.  It must match the name of the CustomResourceDefinition-registration
	// too: plural.group and it must be all lowercase.
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
	// Kind is the serialized kind of the resource.  It is normally CamelCase and singular.
	Kind  string        `json:"kind" protobuf:"bytes,4,opt,name=kind"`
	Scope ResourceScope `json:"scope" protobuf:"bytes,5,opt,name=scope,casttype=ResourceScope"`
}

// ResourceScope is an enum defining the different scopes available to a custom resource
type ResourceScope string

const (
	ClusterScoped   ResourceScope = "Cluster"
	NamespaceScoped ResourceScope = "Namespaced"
)

type Feature struct {
	Trait string `json:"trait" protobuf:"bytes,1,opt,name=trait"`
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

type ProductRef struct {
	VendorID  string `json:"vendorID" protobuf:"bytes,1,opt,name=vendorID"`
	ProductID string `json:"productID" protobuf:"bytes,2,opt,name=productID"`
	PlanID    string `json:"planID" protobuf:"bytes,3,opt,name=planID"`
}

type ResourceDefinitions struct {
	Owned    []ResourceID `json:"owned" protobuf:"bytes,1,rep,name=owned"`
	Required []ResourceID `json:"required" protobuf:"bytes,2,rep,name=required"`
}

// wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available]

type WaitFlags struct {
	Resource     GroupResource         `json:"resource" protobuf:"bytes,1,opt,name=resource"`
	Labels       *metav1.LabelSelector `json:"labels" protobuf:"bytes,2,opt,name=labels"`
	All          bool                  `json:"all" protobuf:"varint,3,opt,name=all"`
	Timeout      metav1.Duration       `json:"timeout" protobuf:"bytes,4,opt,name=timeout"`
	ForCondition string                `json:"for" protobuf:"bytes,5,opt,name=for"`
}

type GroupVersionResource struct {
	Group    string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version  string `json:"version" protobuf:"bytes,2,opt,name=version"`
	Resource string `json:"resource" protobuf:"bytes,3,opt,name=resource"`
}

type GroupResource struct {
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Name  string `json:"name" protobuf:"bytes,2,opt,name=name"`
}
