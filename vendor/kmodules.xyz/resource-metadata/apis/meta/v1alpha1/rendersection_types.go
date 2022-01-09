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
	kmapi "kmodules.xyz/client-go/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindRenderSection = "RenderSection"
	ResourceRenderSection     = "rendersection"
	ResourceRenderSections    = "rendersections"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RenderSection struct {
	metav1.TypeMeta `json:",inline"`
	// Request describes the attributes for the graph request.
	// +optional
	Request *RenderSectionRequest `json:"request,omitempty"`
	// Response describes the attributes for the graph response.
	// +optional
	Response *RenderSectionResponse `json:"response,omitempty"`
}

type RenderSectionRequest struct {
	Source         kmapi.ObjectID  `json:"source"`
	Target         ResourceLocator `json:"target"`
	ConvertToTable bool            `json:"convertToTable"`
}

type RenderSectionResponse struct {
	PageSection `json:",inline"`
}
