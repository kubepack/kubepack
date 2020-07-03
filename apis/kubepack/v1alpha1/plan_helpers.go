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
	"kubepack.dev/kubepack/apis"
	"kubepack.dev/kubepack/crds"

	"k8s.io/apimachinery/pkg/labels"
	"kmodules.xyz/client-go/apiextensions"
)

func (_ Plan) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePlans))
}

func (plan *Plan) SetLabels(planID, prodID, phase string) {
	labelMap := map[string]string{
		apis.LabelPlanID:    planID,
		apis.LabelProductID: prodID,
		apis.LabelPlanPhase: phase,
	}
	plan.ObjectMeta.SetLabels(labelMap)
}

func (_ Plan) FormatLabels(planID, prodID, phase string) string {
	labelMap := make(map[string]string)
	if planID != "" {
		labelMap[apis.LabelPlanID] = planID
	}
	if prodID != "" {
		labelMap[apis.LabelProductID] = prodID
	}
	if phase != "" {
		labelMap[apis.LabelPlanPhase] = phase
	}

	return labels.FormatLabels(labelMap)
}
