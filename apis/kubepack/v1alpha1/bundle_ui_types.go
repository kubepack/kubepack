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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BundleView struct {
	metav1.TypeMeta  `json:",inline"`
	BundleOptionView `json:",inline" protobuf:"bytes,1,opt,name=bundleOptionView"`
	LicenseKey       string `json:"licenseKey,omitempty" protobuf:"bytes,2,opt,name=licenseKey"`
}

type BundleOptionView struct {
	PackageMeta `json:",inline" protobuf:"bytes,1,opt,name=packageMeta"`
	DisplayName string        `json:"displayName" protobuf:"bytes,2,opt,name=displayName"`
	Features    []Feature     `json:"features,omitempty" protobuf:"bytes,3,rep,name=features"`
	Packages    []PackageCard `json:"packages" protobuf:"bytes,4,rep,name=packages"`
}

type PackageCard struct {
	Chart  *ChartCard             `json:"chart,omitempty" protobuf:"bytes,1,opt,name=chart"`
	Bundle *BundleOptionView      `json:"bundle,omitempty" protobuf:"bytes,2,opt,name=bundle"`
	OneOf  *OneOfBundleOptionView `json:"oneOf,omitempty" protobuf:"bytes,3,rep,name=oneOf"`
}

type OneOfBundleOptionView struct {
	Description string              `json:"description" protobuf:"bytes,1,opt,name=description"`
	Bundles     []*BundleOptionView `json:"bundles,omitempty" protobuf:"bytes,2,rep,name=bundles"`
}

type ChartCard struct {
	ChartRef          `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`
	PackageDescriptor `json:",inline" protobuf:"bytes,2,opt,name=packageDescriptor"`
	Features          []string        `json:"features,omitempty" protobuf:"bytes,3,rep,name=features"`
	Namespace         string          `json:"namespace,omitempty" protobuf:"bytes,4,opt,name=namespace"`
	Versions          []VersionOption `json:"versions" protobuf:"bytes,5,rep,name=versions"`
	MultiSelect       bool            `json:"multiSelect,omitempty" protobuf:"varint,6,opt,name=multiSelect"`
	Required          bool            `json:"required,omitempty" protobuf:"varint,7,opt,name=required"`
	Selected          bool            `json:"selected,omitempty" protobuf:"varint,8,opt,name=selected"`
}

type FeatureTable struct {
	Plans []string `json:"plans" protobuf:"bytes,1,rep,name=plans"`
	Rows  []*Row   `json:"rows" protobuf:"bytes,2,rep,name=rows"`
}

type Row struct {
	Trait  string   `json:"trait" protobuf:"bytes,1,opt,name=trait"`
	Values []string `json:"values" protobuf:"bytes,2,rep,name=values"`
}
