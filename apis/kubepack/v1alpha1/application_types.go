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

// ImageSpec contains information about an image used as an icon.
type ImageSpec struct {
	// The source for image represented as either an absolute URL to the image or a Data URL containing
	// the image. Data URLs are defined in RFC 2397.
	Source string `json:"src" protobuf:"bytes,1,opt,name=src"`

	// (optional) The size of the image in pixels (e.g., 25x25).
	TotalSize string `json:"size,omitempty" protobuf:"bytes,2,opt,name=size"`

	// (optional) The mine type of the image (e.g., "image/png").
	Type string `json:"type,omitempty" protobuf:"bytes,3,opt,name=type"`
}

// MediaSpec contains information about an image/video.
type MediaSpec struct {
	// Description is human readable content explaining the purpose of the link.
	Description MediaType `json:"description,omitempty" protobuf:"bytes,1,opt,name=description,casttype=MediaType"`

	ImageSpec `json:",inline" protobuf:"bytes,2,opt,name=imageSpec"`
}

// ContactData contains information about an individual or organization.
type ContactData struct {
	// Name is the descriptive name.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Url could typically be a website address.
	URL string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`

	// Email is the email address.
	Email string `json:"email,omitempty" protobuf:"bytes,3,opt,name=email"`
}

// Link contains information about an URL to surface documentation, dashboards, etc.
type Link struct {
	// Description is human readable content explaining the purpose of the link.
	Description LinkType `json:"description,omitempty" protobuf:"bytes,1,opt,name=description,casttype=LinkType"`

	// Url typically points at a website address.
	URL string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApplicationPackage struct {
	metav1.TypeMeta `json:",inline"`
	Bundle          *ChartRepoRef `json:"bundle,omitempty" protobuf:"bytes,1,opt,name=bundle"`
	Chart           ChartRepoRef  `json:"chart" protobuf:"bytes,2,opt,name=chart"`
	Channel         ChannelType   `json:"channel" protobuf:"bytes,3,opt,name=channel,casttype=ChannelType"`
}

type ChannelType string

const (
	RapidChannel   ChannelType = "Rapid"
	RegularChannel ChannelType = "Regular"
	StableChannel  ChannelType = "Stable"
)
