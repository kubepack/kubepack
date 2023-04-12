/*
Copyright 2023.

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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	releasesv1alpha1 "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

// +kubebuilder:object:generate:=false
type Preset interface {
	GetObjectKind() schema.ObjectKind
	GetName() string
	GetDisplayName() string
	GetLabels() map[string]string
	GetSpec() ClusterChartPresetSpec
}

var _ Preset = &ClusterChartPreset{}

func (in ClusterChartPreset) GetDisplayName() string {
	if in.Spec.DisplayName != "" {
		return in.Spec.DisplayName
	}
	return in.Name
}

func (in ClusterChartPreset) GetSpec() ClusterChartPresetSpec {
	return in.Spec
}

var _ Preset = &ChartPreset{}

func (in ChartPreset) GetDisplayName() string {
	if in.Spec.DisplayName != "" {
		return in.Spec.DisplayName
	}
	return in.Name
}

func (in ChartPreset) GetSpec() ClusterChartPresetSpec {
	return in.Spec
}

type ChartPresetFlatRef struct {
	releasesv1alpha1.ChartSourceFlatRef `json:",inline"`
	PresetGroup                         string `json:"presetGroup,omitempty"`
	PresetKind                          string `json:"presetKind,omitempty"`
	PresetName                          string `json:"presetName,omitempty"`
	PresetSelector                      string `json:"presetSelector,omitempty"`
	Namespace                           string `json:"namespace,omitempty"`
}

func (ref ChartPresetFlatRef) ClusterChartPreset() (*ClusterChartPreset, error) {
	if ref.PresetKind != ResourceKindClusterChartPreset {
		return nil, fmt.Errorf("unknown preset kind %s", ref.PresetKind)
	}

	ps := ClusterChartPreset{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       ResourceKindClusterChartPreset,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "",
		},
		Spec: ClusterChartPresetSpec{
			// Values: nil,
		},
	}

	if ref.PresetName != "" || ref.PresetSelector != "" {
		group := ref.PresetGroup
		if group == "" {
			group = GroupVersion.Group
		}

		presetRef := TypedLocalObjectReference{
			APIGroup: &group,
			Kind:     ref.PresetKind,
			Name:     ref.PresetName,
		}
		if ref.PresetSelector != "" {
			selector, err := metav1.ParseToLabelSelector(ref.PresetSelector)
			if err != nil {
				return nil, err
			}
			presetRef.Selector = selector
		}
		ps.Spec.UsePresets = []TypedLocalObjectReference{presetRef}
	}

	return &ps, nil
}
