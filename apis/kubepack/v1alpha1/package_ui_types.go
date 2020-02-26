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
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type PackageDescriptor struct {
	// Type is the type of the application (e.g. WordPress, MySQL, Cassandra).
	Type string `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`

	// Description is a brief string description of the Application.
	Description string `json:"description,omitempty" protobuf:"bytes,2,opt,name=description"`

	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type.
	Icons []ImageSpec `json:"icons,omitempty" protobuf:"bytes,3,rep,name=icons"`

	// Maintainers is an optional list of maintainers of the application. The maintainers in this list maintain the
	// the source code, images, and package for the application.
	Maintainers []ContactData `json:"maintainers,omitempty" protobuf:"bytes,4,rep,name=maintainers"`

	// Keywords is an optional list of key words associated with the application (e.g. MySQL, RDBMS, database).
	Keywords []string `json:"keywords,omitempty" protobuf:"bytes,5,rep,name=keywords"`

	// Links are a list of descriptive URLs intended to be used to surface additional documentation, dashboards, etc.
	Links []Link `json:"links,omitempty" protobuf:"bytes,6,rep,name=links"`

	// Notes contain a human readable snippets intended as a quick start for the users of the Application.
	// CommonMark markdown syntax may be used for rich text representation.
	Notes string `json:"notes,omitempty" protobuf:"bytes,7,opt,name=notes"`
}

type PackageMeta struct {
	PackageDescriptor `json:",inline" protobuf:"bytes,1,opt,name=packageDescriptor"`

	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
	URL  string `json:"url" protobuf:"bytes,3,opt,name=url"`
	// Version is an optional version indicator for the Application.
	Version string `json:"version,omitempty" protobuf:"bytes,4,opt,name=version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PackageView struct {
	metav1.TypeMeta `json:",inline"`
	PackageMeta     `json:",inline" protobuf:"bytes,1,opt,name=packageMeta"`

	// Default chart values
	Values *runtime.RawExtension `json:"values,omitempty" protobuf:"bytes,2,opt,name=values"`

	// openAPIV3Schema describes the schema used for validation and pruning of the Values file.
	// +optional
	OpenAPIV3Schema *crdv1beta1.JSONSchemaProps `json:"openAPIV3Schema,omitempty" protobuf:"bytes,3,opt,name=openAPIV3Schema"`
}
