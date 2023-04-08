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
	"k8s.io/apimachinery/pkg/util/sets"
	releasesv1alpha1 "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

const (
	ResourceKindPlan = "Plan"
	ResourcePlan     = "plan"
	ResourcePlans    = "plans"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type Plan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PlanSpec   `json:"spec,omitempty"`
	Status            PlanStatus `json:"status,omitempty"`
}

type PlanSpec struct {
	StripeID    string `json:"id"`
	NickName    string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	ProductID   string `json:"productID"`
	Phase       Phase  `json:"phase"`
	// Plans for sorted by weight before displaying to users
	Weight int32 `json:"weight"`

	Bundle *releasesv1alpha1.ChartRef `json:"bundle,omitempty"`
	//+optional
	IncludedPlans []string `json:"includedPlans,omitempty"`

	//+optional
	PricingPattern map[ResourceGroup]PricingPattern `json:"pricingPattern,omitempty"`

	AggregateUsage  *string             `json:"aggregateUsage,omitempty"`
	Amount          *int64              `json:"amount,omitempty"`
	AmountDecimal   *float64            `json:"amountDecimal,string,omitempty"`
	BillingScheme   *string             `json:"billingScheme,omitempty"`
	Currency        *string             `json:"currency,omitempty"`
	Interval        *string             `json:"interval,omitempty"`
	IntervalCount   *int64              `json:"intervalCount,omitempty"`
	Tiers           []*PlanTier         `json:"tiers,omitempty"`
	TiersMode       *string             `json:"tiersMode,omitempty"`
	TransformUsage  *PlanTransformUsage `json:"transformUsage,omitempty"`
	TrialPeriodDays *int64              `json:"trialPeriodDays,omitempty"`
	UsageType       *string             `json:"usageType,omitempty"`
}

type ResourceGroup string

type PricingPattern struct {
	//+optional
	Expression Expression `json:"expression,omitempty"`
	//+optional
	SizedPrices []SizedPrice `json:"sizedPrices,omitempty"`
}

type Expression string

type SizedPrice struct {
	CPU    string  `json:"cpu"`
	Memory string  `json:"memory"`
	Price  float64 `json:"price"`
}

// PlanTier configures tiered pricing
type PlanTier struct {
	FlatAmount        *int64   `json:"flatAmount,omitempty"`
	FlatAmountDecimal *float64 `json:"flatAmountDecimal,string,omitempty"`
	UnitAmount        *int64   `json:"unitAmount,omitempty"`
	UnitAmountDecimal *float64 `json:"unitAmountDecimal,string,omitempty"`
	UpTo              *int64   `json:"upTo,omitempty"`
}

// PlanTransformUsage represents the bucket billing configuration.
type PlanTransformUsage struct {
	DivideBy *int64  `json:"divideBy,omitempty"`
	Round    *string `json:"round,omitempty"`
}

func (p Plan) BundledPlans() []string {
	plans := sets.NewString(p.Spec.StripeID)
	plans.Insert(p.Spec.IncludedPlans...)
	return plans.List()
}

// +kubebuilder:object:root=true

type PlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Plan `json:"items,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Plan{}, &PlanList{})
}

type PlanStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

type FeatureTable struct {
	Plans []Plan `json:"plans"`
	Rows  []*Row `json:"rows"`
}

type Row struct {
	Trait  string   `json:"trait"`
	Values []string `json:"values"`
}
