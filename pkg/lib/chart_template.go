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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/lib-helm/repo"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"gomodules.xyz/jsonpatch/v3"
	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ylib "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"kmodules.xyz/resource-metadata/hub"
	"sigs.k8s.io/application/api/app/v1beta1"
	app_cs "sigs.k8s.io/application/client/clientset/versioned"
	"sigs.k8s.io/yaml"
)

type EditorParameters struct {
	ValuesFile string `json:"valuesFile,omitempty"`
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"valuesPatch,omitempty"`
}

type EditResourceOrder struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`

	ReleaseName string `json:"releaseName"`
	Namespace   string `json:"namespace"`
	Values      string `json:"values"`
}

type EditorOptions struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`

	ReleaseName string                `json:"releaseName"`
	Namespace   string                `json:"namespace"`
	ValuesFile  string                `json:"valuesFile"`
	ValuesPatch *runtime.RawExtension `json:"valuesPatch"`
}

type EditorModel struct {
	Values map[string]interface{} `json:"values"`
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

func LoadEditorModel(cfg *rest.Config, reg *repo.Registry, opts EditorOptions) (string, error) {
	rd, err := hub.NewRegistryOfKnownResources().LoadByGVR(schema.GroupVersionResource{
		Group:    opts.Group,
		Version:  opts.Version,
		Resource: opts.Resource,
	})
	if err != nil {
		return "", err
	}

	chrt, err := reg.GetChart(rd.Spec.UI.Editor.URL, rd.Spec.UI.Editor.Name, rd.Spec.UI.Editor.Version)
	if err != nil {
		return "", err
	}

	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(kc.Discovery()))
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return "", err
	}
	ac, err := app_cs.NewForConfig(cfg)
	if err != nil {
		return "", err
	}

	app, err := ac.AppV1beta1().Applications(opts.Namespace).Get(context.TODO(), opts.ReleaseName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	_, values, err := EditorChartValueManifest(app, mapper, dc, opts.ReleaseName, opts.Namespace, chrt.Chart)
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func EditorChartValueManifest(app *v1beta1.Application, mapper *restmapper.DeferredDiscoveryRESTMapper, dc dynamic.Interface, releaseName, namespace string, chrt *chart.Chart) (string, map[string]interface{}, error) {
	selector, err := metav1.LabelSelectorAsSelector(app.Spec.Selector)
	if err != nil {
		return "", nil, err
	}
	labelSelector := selector.String()

	var buf bytes.Buffer
	values := map[string]interface{}{}

	// detect apiVersion from defaultValues in chart
	gkToVersion := map[metav1.GroupKind]string{}
	for rsKey, x := range chrt.Values {
		var tm metav1.TypeMeta
		err := mapstructure.Decode(x, &tm)
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse TypeMeta for rsKey %s in chart name=%s version=%s values", rsKey, chrt.Name(), chrt.Metadata.Version)
		}
		gv, err := schema.ParseGroupVersion(tm.APIVersion)
		if err != nil {
			return "", nil, err
		}
		gkToVersion[metav1.GroupKind{
			Group: gv.Group,
			Kind:  tm.Kind,
		}] = gv.Version
	}

	for _, gk := range app.Spec.ComponentGroupKinds {
		version, ok := gkToVersion[gk]
		if !ok {
			return "", nil, fmt.Errorf("failed to detect version for GK %#v in chart name=%s version=%s values", gk, chrt.Name(), chrt.Metadata.Version)

		}

		mapping, err := mapper.RESTMapping(schema.GroupKind{
			Group: gk.Group,
			Kind:  gk.Kind,
		}, version)
		if err != nil {
			return "", nil, err
		}
		var rc dynamic.ResourceInterface
		if mapping.Scope == meta.RESTScopeNamespace {
			rc = dc.Resource(mapping.Resource).Namespace(namespace)
		} else {
			rc = dc.Resource(mapping.Resource)
		}

		list, err := rc.List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return "", nil, err
		}
		for _, obj := range list.Items {
			// remove status
			delete(obj.Object, "status")

			buf.WriteString("\n---\n")
			data, err := yaml.Marshal(obj)
			if err != nil {
				return "", nil, err
			}
			buf.Write(data)

			rsKey, err := ResourceKey(obj.GetAPIVersion(), obj.GetKind(), releaseName, obj.GetName())
			if err != nil {
				return "", nil, err
			}
			if _, ok := values[rsKey]; ok {
				return "", nil, fmt.Errorf("duplicate resource key %s for application %s/%s", rsKey, app.Namespace, app.Name)
			}
			values[rsKey] = &obj
		}
	}
	return buf.String(), values, nil
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
		ReleaseName: opts.ReleaseName,
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
		rsKey, err := ResourceKey(obj.GetAPIVersion(), obj.GetKind(), rd.Spec.UI.Options.Name, obj.GetName())
		if err != nil {
			return err
		}

		// values
		modelValues[rsKey] = obj
		return nil
	})
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(modelValues)
	if err != nil {
		panic(err)
	}
	return string(data), err
}

func CreateEditResourceOrder(reg *repo.Registry, opts EditResourceOrder) (*v1alpha1.Order, error) {
	rd, err := hub.NewRegistryOfKnownResources().LoadByGVR(schema.GroupVersionResource{
		Group:    opts.Group,
		Version:  opts.Version,
		Resource: opts.Resource,
	})
	if err != nil {
		return nil, err
	}

	// editor chart
	chrt, err := reg.GetChart(rd.Spec.UI.Editor.URL, rd.Spec.UI.Editor.Name, rd.Spec.UI.Editor.Version)
	if err != nil {
		return nil, err
	}
	originalValues, err := json.Marshal(chrt.Values["values.yaml"])
	if err != nil {
		return nil, err
	}

	modifiedValues, err := ylib.ToJSON([]byte(opts.Values))
	if err != nil {
		return nil, err
	}
	patch, err := jsonpatch.CreatePatch(originalValues, modifiedValues)
	if err != nil {
		return nil, err
	}
	patchData, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	order := v1alpha1.Order{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindOrder,
		}, ObjectMeta: metav1.ObjectMeta{
			Name:              opts.ReleaseName,
			Namespace:         opts.Namespace,
			UID:               types.UID(uuid.New().String()),
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: v1alpha1.OrderSpec{
			Packages: []v1alpha1.PackageSelection{
				{
					Chart: &v1alpha1.ChartSelection{
						ChartRef: v1alpha1.ChartRef{
							URL:  rd.Spec.UI.Editor.URL,
							Name: rd.Spec.UI.Editor.Name,
						},
						Version:     rd.Spec.UI.Editor.Version,
						ReleaseName: opts.ReleaseName,
						Namespace:   opts.Namespace,
						Bundle:      nil,
						ValuesFile:  "values.yaml",
						ValuesPatch: &runtime.RawExtension{
							Raw: patchData,
						},
						Resources: nil,
						WaitFors:  nil,
					},
				},
			},
			KubeVersion: "",
		},
	}
	return &order, err
}
