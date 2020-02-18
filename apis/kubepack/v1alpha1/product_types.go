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
	"k8s.io/apimachinery/pkg/util/sets"
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
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              ProductSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            ProductStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type ProductSpec struct {
	StripeID  string `json:"id" protobuf:"bytes,1,opt,name=id"`
	Key       string `json:"key" protobuf:"bytes,2,opt,name=key"`
	Name      string `json:"name" protobuf:"bytes,3,opt,name=name"`
	ShortName string `json:"shortName" protobuf:"bytes,4,opt,name=shortName,json=shortName"`
	Tagline   string `json:"tagline" protobuf:"bytes,5,opt,name=tagline"`
	//+optional
	Summary string `json:"summary,omitempty" protobuf:"bytes,6,opt,name=summary"`
	Owner   int64  `json:"owner" protobuf:"varint,7,opt,name=owner"`
	Phase   Phase  `json:"phase" protobuf:"bytes,8,opt,name=phase,casttype=Phase"`
	//+optional
	Description string `json:"description,omitempty" protobuf:"bytes,9,opt,name=description"`
	//+optional
	Media []MediaSpec `json:"icons,omitempty" protobuf:"bytes,10,rep,name=icons"`
	//+optional
	Maintainers []ContactData `json:"maintainers,omitempty" protobuf:"bytes,11,rep,name=maintainers"`
	//+optional
	Keywords []string `json:"keywords,omitempty" protobuf:"bytes,12,rep,name=keywords"`
	//+optional
	Links []Link `json:"links,omitempty" protobuf:"bytes,13,rep,name=links"`
	//+optional
	Badges []Badge `json:"badges,omitempty" protobuf:"bytes,14,rep,name=badges"`
	//+optional
	Plans []Plan `json:"plans,omitempty" protobuf:"bytes,15,rep,name=plans"`
	//+optional
	Versions []ProductVersion `json:"versions,omitempty" protobuf:"bytes,16,rep,name=versions"`
	//+optional
	LatestVersion string `json:"latestVersion,omitempty" protobuf:"bytes,17,opt,name=latestVersion"`
	//+optional
	SubProjects map[string]ProjectRef `json:"subProjects,omitempty" protobuf:"bytes,18,rep,name=subProjects"`
}

type Phase string

const (
	PhaseDraft    Phase = "Draft"
	PhaseActive   Phase = "Active"
	PhaseArchived Phase = "Archived"
)

type Plan struct {
	StripeID string   `json:"id" protobuf:"bytes,1,opt,name=id"`
	NickName string   `json:"name" protobuf:"bytes,2,opt,name=name"`
	Chart    ChartRef `json:"chart" protobuf:"bytes,3,opt,name=chart"`
	Phase    Phase    `json:"phase" protobuf:"bytes,4,opt,name=phase,casttype=Phase"`
	//+optional
	IncludedPlans []string `json:"includedPlans,omitempty" protobuf:"bytes,5,rep,name=includedPlans"`
}

func (p Plan) BundledPlans() []string {
	plans := sets.NewString(p.StripeID)
	plans.Insert(p.IncludedPlans...)
	return plans.List()
}

type ProductVersion struct {
	Version  string            `json:"version" protobuf:"bytes,1,opt,name=version"`
	HostDocs bool              `json:"hostDocs" protobuf:"varint,2,opt,name=hostDocs"`
	Show     bool              `json:"show,omitempty" protobuf:"varint,3,opt,name=show"`
	DocsDir  string            `json:"docsDir,omitempty" protobuf:"bytes,4,opt,name=docsDir"` // default: "docs"
	Branch   string            `json:"branch,omitempty" protobuf:"bytes,5,opt,name=branch"`
	Info     map[string]string `json:"info,omitempty" protobuf:"bytes,6,rep,name=info"`
}

type ProjectRef struct {
	Dir      string    `json:"dir" protobuf:"bytes,1,opt,name=dir"`
	Mappings []Mapping `json:"mappings" protobuf:"bytes,2,rep,name=mappings"`
}

type Mapping struct {
	Versions           []string `json:"versions" protobuf:"bytes,1,rep,name=versions"`
	SubProjectVersions []string `json:"subProjectVersions" protobuf:"bytes,2,rep,name=subProjectVersions"`
}

// INFO________________________________________________________________________

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
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Alt  string `json:"alt" protobuf:"bytes,2,opt,name=alt"`
	Logo string `json:"logo" protobuf:"bytes,3,opt,name=logo"`
}

// INFO________________________________________________________________________

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Product `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

type ProductStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
}
