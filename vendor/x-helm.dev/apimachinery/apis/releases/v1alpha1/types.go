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
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

type Feature struct {
	Trait string `json:"trait"`
	Value string `json:"value"`
}

type ResourceDefinitions struct {
	Owned    []metav1.GroupVersionResource `json:"owned"`
	Required []metav1.GroupVersionResource `json:"required"`
}

// wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available]

type WaitFlags struct {
	Resource     metav1.GroupResource  `json:"resource"`
	Labels       *metav1.LabelSelector `json:"labels"`
	All          bool                  `json:"all"`
	Timeout      metav1.Duration       `json:"timeout"`
	ForCondition string                `json:"for"`
}
