/*
Copyright AppsCode Inc. and Contributors

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

type Feature struct {
	Trait string `json:"trait" protobuf:"bytes,1,opt,name=trait"`
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

type ResourceDefinitions struct {
	Owned    []metav1.GroupVersionResource `json:"owned" protobuf:"bytes,1,rep,name=owned"`
	Required []metav1.GroupVersionResource `json:"required" protobuf:"bytes,2,rep,name=required"`
}

// wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available]

type WaitFlags struct {
	Resource     metav1.GroupResource  `json:"resource" protobuf:"bytes,1,opt,name=resource"`
	Labels       *metav1.LabelSelector `json:"labels" protobuf:"bytes,2,opt,name=labels"`
	All          bool                  `json:"all" protobuf:"varint,3,opt,name=all"`
	Timeout      metav1.Duration       `json:"timeout" protobuf:"bytes,4,opt,name=timeout"`
	ForCondition string                `json:"for" protobuf:"bytes,5,opt,name=for"`
}
