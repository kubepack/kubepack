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

package hub

import (
	"context"
	"encoding/json"
	"os"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metadata/apis/shared"

	fluxsrc "github.com/fluxcd/source-controller/api/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	chartsapi "x-helm.dev/apimachinery/apis/charts/v1alpha1"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

const (
	AppsCodeChartsLegacyURL = "https://charts.appscode.com/stable"
	// AppsCodeChartsOCIURL    = "oci://r.appscode.com/charts"
	AppsCodeChartsOCIURL = "oci://ghcr.io/appscode-charts"
	FluxCDChartsURL      = "https://fluxcd-community.github.io/helm-charts"
)

const (
	bootstrapHelmRepositoryName       = "bootstrap"
	EnvVarBootstrapHelmRepositoryName = "BOOTSTRAP_HELM_REPOSITORY_NAME"

	bootstrapHelmRepositoryNamespace       = "kubeops"
	EnvVarBootstrapHelmRepositoryNamespace = "BOOTSTRAP_HELM_REPOSITORY_NAMESPACE"

	BootstrapPresetsName = "bootstrap-presets"

	ChartACE                = "ace"
	ChartACEInstaller       = "ace-installer"
	ChartLicenseProxyServer = "license-proxyserver"
	ChartOpscenterFeatures  = "opscenter-features"

	ChartFluxCD                 = "flux2"
	ChartFluxCDReleaseNamespace = "flux-system"
)

func BootstrapHelmRepositoryNamespace() string {
	ns := os.Getenv(EnvVarBootstrapHelmRepositoryNamespace)
	if ns != "" {
		return ns
	}
	return bootstrapHelmRepositoryNamespace
}

func BootstrapHelmRepositoryName() string {
	ns := os.Getenv(EnvVarBootstrapHelmRepositoryName)
	if ns != "" {
		return ns
	}
	return bootstrapHelmRepositoryName
}

func BootstrapHelmRepository(kc client.Client) kmapi.TypedObjectReference {
	if kc != nil {
		var repo fluxsrc.HelmRepository
		err := kc.Get(context.TODO(), client.ObjectKey{Name: BootstrapHelmRepositoryName(), Namespace: BootstrapHelmRepositoryNamespace()}, &repo)
		if err == nil {
			return kmapi.TypedObjectReference{
				APIGroup:  releasesapi.SourceGroupHelmRepository,
				Kind:      releasesapi.SourceKindHelmRepository,
				Namespace: BootstrapHelmRepositoryNamespace(),
				Name:      BootstrapHelmRepositoryName(),
			}
		}
	}
	return kmapi.TypedObjectReference{
		APIGroup:  releasesapi.SourceGroupLegacy,
		Kind:      releasesapi.SourceKindLegacy,
		Namespace: "",
		Name:      AppsCodeChartsLegacyURL,
	}
}

func FluxCDHelmRepository(kc client.Client) kmapi.TypedObjectReference {
	if kc != nil {
		var repo fluxsrc.HelmRepository
		err := kc.Get(context.TODO(), client.ObjectKey{Name: BootstrapHelmRepositoryName(), Namespace: BootstrapHelmRepositoryNamespace()}, &repo)
		if err == nil {
			return kmapi.TypedObjectReference{
				APIGroup:  releasesapi.SourceGroupHelmRepository,
				Kind:      releasesapi.SourceKindHelmRepository,
				Namespace: BootstrapHelmRepositoryNamespace(),
				Name:      BootstrapHelmRepositoryName(),
			}
		}
	}
	return kmapi.TypedObjectReference{
		APIGroup:  releasesapi.SourceGroupLegacy,
		Kind:      releasesapi.SourceKindLegacy,
		Namespace: "",
		Name:      FluxCDChartsURL,
	}
}

func GetBootstrapPresets(kc client.Client) (*shared.BootstrapPresets, bool) {
	if kc != nil {
		var ccp chartsapi.ClusterChartPreset
		err := kc.Get(context.TODO(), client.ObjectKey{Name: BootstrapPresetsName}, &ccp)
		if err == nil {
			var preset shared.BootstrapPresets
			err := json.Unmarshal(ccp.Spec.Values.Raw, &preset)
			if err == nil {
				return &preset, true
			}
		}
	}
	return nil, false
}

func FeatureVersion(kc client.Client, featureName string) string {
	preset, found := GetBootstrapPresets(kc)
	if found {
		hr := preset.Helm.Releases[featureName]
		if hr != nil {
			return hr.Version
		}
	}
	return ""
}

func FeatureValues(kc client.Client, featureName string) (map[string]any, error) {
	preset, found := GetBootstrapPresets(kc)
	if found {
		hr := preset.Helm.Releases[featureName]
		if hr != nil && hr.Values != nil {
			var vals map[string]any
			if err := toMap(hr.Values, &vals); err != nil {
				return nil, err
			}
			return vals, nil
		}
	}
	return map[string]any{}, nil
}

func toMap(src, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func HelmCreateNamespace(kc client.Client) bool {
	preset, found := GetBootstrapPresets(kc)
	if found {
		return preset.Helm.CreateNamespace
	}
	return true
}

func IsFeaturesetGR(gr schema.GroupResource) bool {
	return gr.Group == "ui.k8s.appscode.com" && gr.Resource == "featuresets"
}

func IsFeaturesetGK(gk schema.GroupKind) bool {
	return gk.Group == "ui.k8s.appscode.com" && gk.Kind == "FeatureSet"
}
