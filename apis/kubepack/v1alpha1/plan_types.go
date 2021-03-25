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
)

const (
	ResourceKindPlan = "Plan"
	ResourcePlan     = "plan"
	ResourcePlans    = "plans"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=plans,singular=plan,scope=Cluster,categories={kubepack,appscode}
// +kubebuilder:subresource:status
type Plan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              PlanSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            PlanStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type PlanSpec struct {
	StripeID    string `json:"id" protobuf:"bytes,1,opt,name=id"`
	NickName    string `json:"name" protobuf:"bytes,2,opt,name=name"`
	DisplayName string `json:"displayName" protobuf:"bytes,3,opt,name=displayName"`
	Description string `json:"description" protobuf:"bytes,4,opt,name=description"`
	ProductID   string `json:"productID" protobuf:"bytes,5,opt,name=productID"`
	Phase       Phase  `json:"phase" protobuf:"bytes,6,opt,name=phase,casttype=Phase"`
	// Plans for sorted by weight before displaying to users
	Weight int32 `json:"weight" protobuf:"varint,7,opt,name=weight"`

	Bundle *ChartRef `json:"bundle,omitempty" protobuf:"bytes,8,opt,name=bundle"`
	//+optional
	IncludedPlans []string `json:"includedPlans,omitempty" protobuf:"bytes,9,rep,name=includedPlans"`

	//+optional
	PricingPattern map[ResourceGroup]PricingPattern `json:"pricingPattern,omitempty" protobuf:"bytes,10,rep,name=pricingPattern,castkey=ResourceGroup"`

	AggregateUsage  *string             `json:"aggregateUsage,omitempty" protobuf:"bytes,11,opt,name=aggregateUsage"`
	Amount          *int64              `json:"amount,omitempty" protobuf:"varint,12,opt,name=amount"`
	AmountDecimal   *float64            `json:"amountDecimal,string,omitempty" protobuf:"fixed64,13,opt,name=amountDecimal"`
	BillingScheme   *string             `json:"billingScheme,omitempty" protobuf:"bytes,14,opt,name=billingScheme"`
	Currency        *string             `json:"currency,omitempty" protobuf:"bytes,15,opt,name=currency"`
	Interval        *string             `json:"interval,omitempty" protobuf:"bytes,16,opt,name=interval"`
	IntervalCount   *int64              `json:"intervalCount,omitempty" protobuf:"varint,17,opt,name=intervalCount"`
	Tiers           []*PlanTier         `json:"tiers,omitempty" protobuf:"bytes,18,rep,name=tiers"`
	TiersMode       *string             `json:"tiersMode,omitempty" protobuf:"bytes,19,opt,name=tiersMode"`
	TransformUsage  *PlanTransformUsage `json:"transformUsage,omitempty" protobuf:"bytes,20,opt,name=transformUsage"`
	TrialPeriodDays *int64              `json:"trialPeriodDays,omitempty" protobuf:"varint,21,opt,name=trialPeriodDays"`
	UsageType       *string             `json:"usageType,omitempty" protobuf:"bytes,22,opt,name=usageType"`
}

type ResourceGroup string

type PricingPattern struct {
	//+optional
	Expression Expression `json:"expression,omitempty" protobuf:"bytes,1,opt,name=expression,casttype=Expression"`
	//+optional
	SizedPrices []SizedPrice `json:"sizedPrices,omitempty" protobuf:"bytes,2,rep,name=sizedPrices"`
}

type Expression string

type SizedPrice struct {
	CPU    string  `json:"cpu" protobuf:"bytes,1,opt,name=cpu"`
	Memory string  `json:"memory" protobuf:"bytes,2,opt,name=memory"`
	Price  float64 `json:"price" protobuf:"fixed64,3,opt,name=price"`
}

// PlanTier configures tiered pricing
type PlanTier struct {
	FlatAmount        *int64   `json:"flatAmount,omitempty" protobuf:"varint,1,opt,name=flatAmount"`
	FlatAmountDecimal *float64 `json:"flatAmountDecimal,string,omitempty" protobuf:"fixed64,2,opt,name=flatAmountDecimal"`
	UnitAmount        *int64   `json:"unitAmount,omitempty" protobuf:"varint,3,opt,name=unitAmount"`
	UnitAmountDecimal *float64 `json:"unitAmountDecimal,string,omitempty" protobuf:"fixed64,4,opt,name=unitAmountDecimal"`
	UpTo              *int64   `json:"upTo,omitempty" protobuf:"varint,5,opt,name=upTo"`
}

// PlanTransformUsage represents the bucket billing configuration.
type PlanTransformUsage struct {
	DivideBy *int64  `json:"divideBy,omitempty" protobuf:"varint,1,opt,name=divideBy"`
	Round    *string `json:"round,omitempty" protobuf:"bytes,2,opt,name=round"`
}

func (p Plan) BundledPlans() []string {
	plans := sets.NewString(p.Spec.StripeID)
	plans.Insert(p.Spec.IncludedPlans...)
	return plans.List()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type PlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Plan `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

type PlanStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
}
