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
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceKindOrder = "Order"
	ResourceOrder     = "order"
	ResourceOrders    = "orders"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=orders,singular=order,scope=Cluster,categories={kubepack,appscode}
// +kubebuilder:subresource:status
type Order struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OrderSpec   `json:"spec,omitempty"`
	Status            OrderStatus `json:"status,omitempty"`
}

type OrderSpec struct {
	Packages []PackageSelection `json:"items"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string `json:"kubeVersion,omitempty"`
}

type PackageSelection struct {
	Chart *ChartSelection `json:"chart,omitempty"`
}

type ChartSelection struct {
	ChartRef    `json:",inline"`
	Version     string `json:"version"`
	ReleaseName string `json:"releaseName"`
	Namespace   string `json:"namespace"`

	// Bundle points to the parent bundle which was used to install this, if any.
	// If the chart was installed directly it will be unset.
	Bundle *ChartRepoRef `json:"bundle,omitempty"`

	ValuesFile string `json:"valuesFile,omitempty"`
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"valuesPatch,omitempty"`
	Resources   *ResourceDefinitions  `json:"resources,omitempty"`
	WaitFors    []WaitFlags           `json:"waitFors,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type OrderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Order `json:"items,omitempty"`
}

type OrderStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}
