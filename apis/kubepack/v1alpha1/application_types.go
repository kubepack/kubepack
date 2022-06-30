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

// ImageSpec contains information about an image used as an icon.
type ImageSpec struct {
	// The source for image represented as either an absolute URL to the image or a Data URL containing
	// the image. Data URLs are defined in RFC 2397.
	Source string `json:"src"`

	// (optional) The size of the image in pixels (e.g., 25x25).
	TotalSize string `json:"size,omitempty"`

	// (optional) The mine type of the image (e.g., "image/png").
	Type string `json:"type,omitempty"`
}

// MediaSpec contains information about an image/video.
type MediaSpec struct {
	// Description is human readable content explaining the purpose of the link.
	Description MediaType `json:"description,omitempty"`

	ImageSpec `json:",inline"`
}

// ContactData contains information about an individual or organization.
type ContactData struct {
	// Name is the descriptive name.
	Name string `json:"name,omitempty"`

	// Url could typically be a website address.
	URL string `json:"url,omitempty"`

	// Email is the email address.
	Email string `json:"email,omitempty"`
}

// Link contains information about an URL to surface documentation, dashboards, etc.
type Link struct {
	// Description is human readable content explaining the purpose of the link.
	Description LinkType `json:"description,omitempty"`

	// Url typically points at a website address.
	URL string `json:"url,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApplicationPackage struct {
	metav1.TypeMeta `json:",inline"`
	Bundle          *ChartRepoRef `json:"bundle,omitempty"`
	Chart           ChartRepoRef  `json:"chart"`
	Channel         ChannelType   `json:"channel"`
}

type ChannelType string

const (
	RapidChannel   ChannelType = "Rapid"
	RegularChannel ChannelType = "Regular"
	StableChannel  ChannelType = "Stable"
)
