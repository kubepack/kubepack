/*
Copyright The Helm Authors.

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

package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gobuffalo/flect"
	"gomodules.xyz/sets"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	rspb "helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"kmodules.xyz/client-go/tools/parser"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	driversapi "x-helm.dev/apimachinery/apis/drivers/v1alpha1"
	"x-helm.dev/apimachinery/apis/shared"
)

var empty = struct{}{}

// newAppReleaseSecretsObject constructs a kubernetes AppRelease object
// to store a release. Each configmap data entry is the base64
// encoded gzipped string of a release.
//
// The following labels are used within each configmap:
//
//	"modifiedAt"     - timestamp indicating when this configmap was last modified. (set in Update)
//	"createdAt"      - timestamp indicating when this configmap was created. (set in Create)
//	"version"        - version of the release.
//	"status"         - status of the release (see pkg/release/status.go for variants)
//	"owner"          - owner of the configmap, currently "helm".
//	"name"           - name of the release.
func newAppReleaseObject(rls *rspb.Release) *driversapi.AppRelease {
	const owner = "helm"

	/*
		LabelChartFirstDeployed = "first-deployed.meta.helm.sh/"
		LabelChartLastDeployed  = "last-deployed.meta.helm.sh/"
	*/

	appName := rls.Name
	if partOf, ok := rls.Chart.Metadata.Annotations["app.kubernetes.io/part-of"]; ok {
		appName = partOf
	}

	// create and return configmap object
	obj := &driversapi.AppRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: rls.Namespace,
			Labels: map[string]string{
				"owner": owner,
				fmt.Sprintf("%s/%s", labelScopeReleaseName, rls.Name): rls.Name,
			},
		},
		Spec: driversapi.AppReleaseSpec{
			Descriptor: driversapi.Descriptor{
				Type:        rls.Chart.Metadata.Type,
				Version:     rls.Chart.Metadata.AppVersion,
				Description: rls.Info.Description,
				Owners:      nil, // FIX
				Keywords:    rls.Chart.Metadata.Keywords,
				Links: []shared.Link{
					{
						Description: "website",
						URL:         rls.Chart.Metadata.Home,
					},
				},
				Notes: rls.Info.Notes,
			},
			Release: driversapi.ReleaseInfo{
				Name:          rls.Name,
				Version:       strconv.Itoa(rls.Version),
				Status:        string(rls.Info.Status),
				FirstDeployed: &metav1.Time{Time: rls.Info.FirstDeployed.Time.UTC()},
				LastDeployed:  &metav1.Time{Time: rls.Info.LastDeployed.Time.UTC()},
			},
			Components: nil,
			Selector:   nil,
			// ResourceKeys:
		},
	}
	if rls.Chart.Metadata.Icon != "" {
		var imgType string
		if resp, err := http.Get(rls.Chart.Metadata.Icon); err == nil {
			if mime, err := mimetype.DetectReader(resp.Body); err == nil {
				imgType = mime.String()
			}
			_ = resp.Body.Close()
		}
		obj.Spec.Descriptor.Icons = []shared.ImageSpec{
			{
				Source: rls.Chart.Metadata.Icon,
				// TotalSize: "",
				Type: imgType,
			},
		}
	}
	for _, maintainer := range rls.Chart.Metadata.Maintainers {
		obj.Spec.Descriptor.Maintainers = append(obj.Spec.Descriptor.Maintainers, shared.ContactData{
			Name:  maintainer.Name,
			URL:   maintainer.URL,
			Email: maintainer.Email,
		})
	}

	//lbl := map[string]string{
	//	"app.kubernetes.io/managed-by": "Helm",
	//}
	lbl := make(map[string]string)
	if partOf, ok := rls.Chart.Metadata.Annotations["app.kubernetes.io/part-of"]; ok && partOf != "" {
		lbl["app.kubernetes.io/part-of"] = partOf
	} else {
		lbl["app.kubernetes.io/instance"] = rls.Name

		// ref : https://github.com/kubepack/helm/blob/ac-1.21.0/pkg/action/validate.go#L208-L214
		if data, ok := rls.Chart.Metadata.Annotations["meta.x-helm.dev/editor"]; ok && data != "" {
			var gvr metav1.GroupVersionResource
			if err := json.Unmarshal([]byte(data), &gvr); err == nil {
				lbl["app.kubernetes.io/name"] = fmt.Sprintf("%s.%s", gvr.Resource, gvr.Group)
			}
		}
	}
	obj.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: lbl,
	}

	if editorGVR, ok := rls.Chart.Metadata.Annotations["meta.x-helm.dev/editor"]; ok {
		var gvr metav1.GroupVersionResource
		if err := json.Unmarshal([]byte(editorGVR), &gvr); err != nil {
			panic(err)
		} else {
			obj.Spec.Editor = &gvr
		}

		if f, ok := rls.Config["form"]; ok {
			if fd, err := json.Marshal(f); err == nil {
				obj.Spec.Release.Form = &runtime.RawExtension{Raw: fd}
			}
		}

		if resources, ok, err := unstructured.NestedMap(rls.Chart.Values, "resources"); err == nil && ok {
			resourceKeys := make([]string, 0, len(resources))
			for k := range resources {
				resourceKeys = append(resourceKeys, k)
			}
			sort.Strings(resourceKeys)
			obj.Spec.ResourceKeys = resourceKeys
		}
	}

	components, _, err := parser.ExtractComponentGVKs([]byte(rls.Manifest))
	if err != nil {
		// WARNING(tamal): This error should never happen
		panic(err)
	}

	if data, ok := rls.Chart.Metadata.Annotations["meta.x-helm.dev/resources"]; ok && data != "" {
		var gvks []metav1.GroupVersionKind
		err := yaml.Unmarshal([]byte(data), &gvks)
		if err != nil {
			panic(err)
		}
		for _, gk := range gvks {
			components[gk] = empty
		}
	}
	gvks := make([]metav1.GroupVersionKind, 0, len(components))
	for gvk := range components {
		gvks = append(gvks, gvk)
	}
	sort.Slice(gvks, func(i, j int) bool {
		if gvks[i].Group == gvks[j].Group {
			return gvks[i].Kind < gvks[j].Kind
		}
		return gvks[i].Group < gvks[j].Group
	})
	obj.Spec.Components = gvks

	return obj
}

// decodeRelease decodes the bytes of data into a release
// type. Data must contain a base64 encoded gzipped string of a
// valid release, otherwise an error is returned.
func decodeReleaseFromApp(kc client.Client, app *driversapi.AppRelease, rlsNames []string) ([]*rspb.Release, error) {
	if len(rlsNames) == 0 {
		rlsNames = relevantReleases(app.Labels)
	}

	releases := make([]*rspb.Release, 0, len(rlsNames))

	for range rlsNames {
		var rls rspb.Release

		rls.Name = app.Spec.Release.Name
		rls.Namespace = app.Namespace
		rls.Version, _ = strconv.Atoi(app.Spec.Release.Version)

		// This is not needed or used from release
		//chartURL, ok := app.Annotations[apis.LabelChartURL]
		//if !ok {
		//	return nil, fmt.Errorf("missing %s annotation on AppRelease %s/%s", apis.LabelChartURL, app.Namespace, app.Name)
		//}
		//chartName, ok := app.Annotations[apis.LabelChartName]
		//if !ok {
		//	return nil, fmt.Errorf("missing %s annotation on AppRelease %s/%s", apis.LabelChartName, app.Namespace, app.Name)
		//}
		//chartVersion, ok := app.Annotations[apis.LabelChartVersion]
		//if !ok {
		//	return nil, fmt.Errorf("missing %s annotation on AppRelease %s/%s", apis.LabelChartVersion, app.Namespace, app.Name)
		//}
		//chrt, err := lib.DefaultRegistry.GetChart(chartURL, chartName, chartVersion)
		//if err != nil {
		//	return nil, err
		//}
		//rls.Chart = chrt.Chart

		rls.Info = &release.Info{
			Description: app.Spec.Descriptor.Description,
			Status:      release.Status(app.Spec.Release.Status),
			Notes:       app.Spec.Descriptor.Notes,
		}
		rls.Info.FirstDeployed = helmtime.Time{Time: app.Spec.Release.FirstDeployed.Time}
		rls.Info.LastDeployed = helmtime.Time{Time: app.Spec.Release.LastDeployed.Time}

		editorGVR := app.Spec.Editor
		rlm := types.NamespacedName{
			Name:      rls.Name,
			Namespace: rls.Namespace,
		}
		tpl, err := EditorChartValueManifest(kc, app, rlm, editorGVR)
		if err != nil {
			return nil, err
		}

		rls.Manifest = string(tpl.Manifest)

		if editorGVR != nil {
			rls.Chart = &chart.Chart{
				Values: map[string]interface{}{},
			}
			if app.Spec.Release.Form != nil {
				if form, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app.Spec.Release.Form); err == nil {
					tpl.Values.Object["form"] = form
				}
			}
			rls.Chart.Values = tpl.Values.Object
			rls.Config = tpl.Values.Object
		}
		// else
		// keep tls.Chart nil and see if that causes panics
		// we don't want to load chart from remote here any more, because we want to embed chart in Go binary

		releases = append(releases, &rls)
	}

	return releases, nil
}

func EditorChartValueManifest(kc client.Client, app *driversapi.AppRelease, rls types.NamespacedName, editorGVR *metav1.GroupVersionResource) (*EditorTemplate, error) {
	labels := app.Spec.Selector
	labels.MatchLabels["app.kubernetes.io/instance"] = rls.Name

	selector, err := metav1.LabelSelectorAsSelector(labels)
	if err != nil {
		return nil, err
	}
	// labelSelector := selector.String()

	var buf bytes.Buffer
	resourceMap := map[string]interface{}{}
	resourceKeys := sets.NewString(app.Spec.ResourceKeys...)

	for _, gvk := range app.Spec.Components {
		var list unstructured.UnstructuredList
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind,
		})
		err = kc.List(context.TODO(), &list, client.InNamespace(rls.Namespace), client.MatchingLabelsSelector{Selector: selector})
		if err != nil {
			return nil, err
		}
		for _, obj := range list.Items {
			rsKey, err := ResourceKey(obj.GetAPIVersion(), obj.GetKind(), rls.Name, obj.GetName())
			if err != nil {
				return nil, err
			}

			// skip ownership check for ui-wizards
			// https://github.com/kubepack/helm/blob/ac-1.21.0/pkg/action/validate.go#L87-L92
			if editorGVR == nil {
				annotationMap := obj.GetAnnotations()
				if v := annotationMap["meta.helm.sh/release-name"]; v != rls.Name {
					continue
				}
				if v := annotationMap["meta.helm.sh/release-namespace"]; v != rls.Namespace {
					continue
				}
			} else {
				if resourceKeys.Len() > 0 && !resourceKeys.Has(rsKey) {
					continue
				}
			}

			// remove status
			delete(obj.Object, "status")

			buf.WriteString("\n---\n")
			data, err := yaml.Marshal(&obj)
			if err != nil {
				return nil, err
			}
			buf.Write(data)

			if _, ok := resourceMap[rsKey]; ok {
				return nil, fmt.Errorf("duplicate resource key %s for AppRelease %s/%s", rsKey, app.Namespace, app.Name)
			}
			resourceMap[rsKey] = &obj
		}
	}

	tpl := EditorTemplate{
		Manifest: buf.Bytes(),
	}
	if editorGVR != nil {
		tpl.Values = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"resource": map[string]interface{}{
						"group":    editorGVR.Group,
						"version":  editorGVR.Version,
						"resource": editorGVR.Resource,
					},
					"release": map[string]interface{}{
						"name":      rls.Name,
						"namespace": rls.Namespace,
					},
				},
				"resources": resourceMap,
			},
		}
	}

	return &tpl, nil
}

type EditorTemplate struct {
	Manifest []byte                     `json:"manifest,omitempty"`
	Values   *unstructured.Unstructured `json:"values,omitempty"`
}

func ResourceKey(apiVersion, kind, chartName, name string) (string, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return "", err
	}

	groupPrefix := gv.Group
	groupPrefix = strings.TrimSuffix(groupPrefix, ".k8s.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".kubernetes.io")
	// groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
	groupPrefix = strings.Replace(groupPrefix, ".", "_", -1)
	groupPrefix = flect.Pascalize(groupPrefix)

	var nameSuffix string
	nameSuffix = strings.TrimPrefix(name, chartName)
	// we can't use - as separator since Go template does not like it
	// Go template throws an error like "unexpected bad character U+002D '-' in with"
	// ref: https://github.com/gohugoio/hugo/issues/1474
	nameSuffix = flect.Underscore(nameSuffix)
	nameSuffix = strings.Trim(nameSuffix, "_")

	result := flect.Camelize(groupPrefix + kind)
	if len(nameSuffix) > 0 {
		result += "_" + nameSuffix
	}
	return result, nil
}

func ResourceFilename(apiVersion, kind, chartName, name string) (string, string, string) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		panic(err)
	}

	groupPrefix := gv.Group
	groupPrefix = strings.TrimSuffix(groupPrefix, ".k8s.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".kubernetes.io")
	// groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
	groupPrefix = strings.Replace(groupPrefix, ".", "_", -1)
	groupPrefix = flect.Pascalize(groupPrefix)

	var nameSuffix string
	nameSuffix = strings.TrimPrefix(name, chartName)
	nameSuffix = strings.Replace(nameSuffix, ".", "-", -1)
	nameSuffix = strings.Trim(nameSuffix, "-")
	nameSuffix = flect.Pascalize(nameSuffix)

	return flect.Underscore(kind), flect.Underscore(kind + nameSuffix), flect.Underscore(groupPrefix + kind + nameSuffix)
}
