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

package lib

import (
	"bytes"
	"fmt"
	"strings"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/lib-helm/repo"

	"github.com/gobuffalo/flect"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kmodules.xyz/resource-metadata/hub"
	"sigs.k8s.io/yaml"
)

type EditorParameters struct {
	ValuesFile string `json:"valuesFile,omitempty"`
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"valuesPatch,omitempty"`
}

type EditorOptions struct {
	Group    string
	Version  string
	Resource string

	// Release
	Namespace   string
	ValuesFile  string
	ValuesPatch *runtime.RawExtension
}

type ChartTemplate struct {
	v1alpha1.ChartRef `json:",inline"`
	Version           string                       `json:"version"`
	ReleaseName       string                       `json:"releaseName"`
	Namespace         string                       `json:"namespace"`
	CRDs              []BucketJsonFile             `json:"crds"`
	Manifest          *BucketObject                `json:"manifest"`
	Resources         []*unstructured.Unstructured `json:"resources"`
}

func RenderChartTemplate(bs *BlobStore, reg *repo.Registry, order v1alpha1.Order) (string, []ChartTemplate, error) {
	var buf bytes.Buffer
	var tpls []ChartTemplate

	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		f1 := &TemplateRenderer{
			Registry:    reg,
			ChartRef:    pkg.Chart.ChartRef,
			Version:     pkg.Chart.Version,
			ReleaseName: pkg.Chart.ReleaseName,
			Namespace:   pkg.Chart.Namespace,
			KubeVersion: "v1.17.0",
			ValuesFile:  pkg.Chart.ValuesFile,
			ValuesPatch: pkg.Chart.ValuesPatch,
			BucketURL:   bs.Bucket,
			UID:         string(order.UID),
			PublicURL:   bs.Host,
		}
		err := f1.Do()
		if err != nil {
			return "", nil, err
		}

		tpl := ChartTemplate{
			ChartRef:    pkg.Chart.ChartRef,
			Version:     pkg.Chart.Version,
			ReleaseName: pkg.Chart.ReleaseName,
			Namespace:   pkg.Chart.Namespace,
		}
		crds, manifestFile := f1.Result()
		for _, crd := range crds {
			resources, err := ExtractResources(crd.Data)
			if err != nil {
				return "", nil, err
			}
			if len(resources) != 1 {
				return "", nil, fmt.Errorf("%d crds found in %s", len(resources), crd.Filename)
			}
			tpl.CRDs = append(tpl.CRDs, BucketJsonFile{
				URL:      crd.URL,
				Key:      crd.Key,
				Filename: crd.Filename,
				Data:     resources[0],
			})
		}
		if manifestFile != nil {
			tpl.Manifest = &BucketObject{
				URL: manifestFile.URL,
				Key: manifestFile.Key,
			}
			tpl.Resources, err = ExtractResources(manifestFile.Data)
			if err != nil {
				return "", nil, err
			}
			_, err = fmt.Fprintf(&buf, "---\n# Source: %s - %s@%s\n", f1.ChartRef.URL, f1.ChartRef.Name, f1.Version)
			if err != nil {
				return "", nil, err
			}

			_, err := buf.Write(manifestFile.Data)
			if err != nil {
				return "", nil, err
			}
			_, err = buf.WriteRune('\n')
			if err != nil {
				return "", nil, err
			}
		}
		tpls = append(tpls, tpl)
	}

	return buf.String(), tpls, nil
}

func GenerateEditorModel(reg *repo.Registry, opts EditorOptions) (string, error) {
	rd, err := hub.NewRegistryOfKnownResources().LoadByGVR(schema.GroupVersionResource{
		Group:    opts.Group,
		Version:  opts.Version,
		Resource: opts.Resource,
	})
	if err != nil {
		return "", err
	}

	f1 := &EditorModelGenerator{
		Registry: reg,
		ChartRef: v1alpha1.ChartRef{
			URL:  rd.Spec.UI.Options.URL,
			Name: rd.Spec.UI.Options.Name,
		},
		Version:     rd.Spec.UI.Options.Version,
		ReleaseName: rd.Spec.UI.Options.Name,
		Namespace:   opts.Namespace,
		KubeVersion: "v1.17.0",
		ValuesFile:  opts.ValuesFile,
		ValuesPatch: opts.ValuesPatch,
	}
	err = f1.Do()
	if err != nil {
		return "", err
	}

	modelValues := map[string]*unstructured.Unstructured{}
	err = ProcessResources(f1.Result(), func(obj *unstructured.Unstructured) error {
		rsKey, err := resourceKey(obj.GetAPIVersion(), obj.GetKind(), rd.Spec.UI.Options.Name, obj.GetName())
		if err != nil {
			return err
		}

		// values
		modelValues[rsKey] = obj
		return nil
	})

	data, err := yaml.Marshal(modelValues)
	if err != nil {
		panic(err)
	}
	return string(data), err
}

func resourceKey(apiVersion, kind, chartName, name string) (string, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return "", err
	}

	groupPrefix := gv.Group
	groupPrefix = strings.TrimSuffix(groupPrefix, ".k8s.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".kubernetes.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
	groupPrefix = strings.Replace(groupPrefix, ".", "_", -1)
	groupPrefix = flect.Pascalize(groupPrefix)

	var nameSuffix string
	nameSuffix = strings.TrimPrefix(chartName, name)
	nameSuffix = strings.Replace(nameSuffix, ".", "-", -1)
	nameSuffix = strings.Trim(nameSuffix, "-")
	nameSuffix = flect.Pascalize(nameSuffix)

	return flect.Camelize(groupPrefix + kind + nameSuffix), nil
}
