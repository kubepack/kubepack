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
	"x-helm.dev/apimachinery/apis/shared"
)

const (
	ResourceKindProduct = "Product"
	ResourceProduct     = "product"
	ResourceProducts    = "products"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
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
	Media []shared.MediaSpec `json:"icons,omitempty"`
	//+optional
	Maintainers []shared.ContactData `json:"maintainers,omitempty"`
	//+optional
	Keywords []string `json:"keywords,omitempty"`
	//+optional
	Links []shared.Link `json:"links,omitempty"`
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

type Badge struct {
	URL  string `json:"url"`
	Alt  string `json:"alt"`
	Logo string `json:"logo"`
}

// +kubebuilder:object:root=true

type ProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Product `json:"items,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Product{}, &ProductList{})
}

type ProductStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}
