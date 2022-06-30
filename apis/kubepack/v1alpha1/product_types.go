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

const (
	ResourceKindProduct = "Product"
	ResourceProduct     = "product"
	ResourceProducts    = "products"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=products,singular=product,scope=Cluster,categories={kubepack,appscode}
// +kubebuilder:subresource:status
type Product struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProductSpec   `json:"spec,omitempty"`
	Status            ProductStatus `json:"status,omitempty"`
}

type ProductSpec struct {
	StripeID  string `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Tagline   string `json:"tagline"`
	//+optional
	Summary string `json:"summary,omitempty"`
	Owner   int64  `json:"owner"`
	//+optional
	OwnerName string `json:"ownerName,omitempty"`
	//+optional
	Description string `json:"description,omitempty"`
	//+optional
	UnitLabel string `json:"unitLabel,omitempty"`
	Phase     Phase  `json:"phase"`
	//+optional
	Media []MediaSpec `json:"icons,omitempty"`
	//+optional
	Maintainers []ContactData `json:"maintainers,omitempty"`
	//+optional
	Keywords []string `json:"keywords,omitempty"`
	//+optional
	Links []Link `json:"links,omitempty"`
	//+optional
	Badges []Badge `json:"badges,omitempty"`
	//+optional
	Versions []ProductVersion `json:"versions,omitempty"`
	//+optional
	LatestVersion string `json:"latestVersion,omitempty"`
}

type Phase string

const (
	PhaseDraft    Phase = "Draft"
	PhaseActive   Phase = "Active"
	PhaseArchived Phase = "Archived"
)

type ProductVersion struct {
	Version string `json:"version"`
	// +optional
	ReleaseDate *metav1.Time `json:"releaseDate,omitempty"`
}

type MediaType string

const (
	MediaLogo        MediaType = "logo"
	MediaLogoWhite   MediaType = "logo_white"
	MediaIcon        MediaType = "icon"
	MediaIcon192_192 MediaType = "icon_192x192"
	MediaHeroImage   MediaType = "hero_image"
	MediaIntroVideo  MediaType = "intro_video"
)

type LinkType string

const (
	LinkWebsite         LinkType = "website"
	LinkSupportDesk     LinkType = "support_desk"
	LinkFacebook        LinkType = "facebook"
	LinkLinkedIn        LinkType = "linkedin"
	LinkTwitter         LinkType = "twitter"
	LinkTwitterID       LinkType = "twitter_id"
	LinkYouTube         LinkType = "youtube"
	LinkSourceRepo      LinkType = "src_repo"
	LinkStarRepo        LinkType = "star_repo"
	LinkDocsRepo        LinkType = "docs_repo"
	LinkDatasheetFormID LinkType = "datasheet_form_id"
)

type Badge struct {
	URL  string `json:"url"`
	Alt  string `json:"alt"`
	Logo string `json:"logo"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Product `json:"items,omitempty"`
}

type ProductStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}
