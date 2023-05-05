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
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gobuffalo/flect"
	"gomodules.xyz/sets"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	rspb "helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"kmodules.xyz/client-go/apiextensions"
	disco_util "kmodules.xyz/client-go/discovery"
	"kmodules.xyz/client-go/tools/parser"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	driversapi "x-helm.dev/apimachinery/apis/drivers/v1alpha1"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
	"x-helm.dev/apimachinery/apis/shared"
)

var empty = struct{}{}

func mustNewAppReleaseObject(rls *rspb.Release) *driversapi.AppRelease {
	out, err := newAppReleaseObject(rls)
	if err != nil {
		panic(err)
	}
	return out
}

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
func newAppReleaseObject(rls *rspb.Release) (*driversapi.AppRelease, error) {
	const owner = "helm"

	appName := rls.Name
	if partOf, ok := rls.Chart.Metadata.Annotations["app.kubernetes.io/part-of"]; ok {
		appName = partOf
	}

	// create and return configmap object
	obj := &driversapi.AppRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: rls.Namespace,
			// labels are required by https://github.com/x-helm/helm/blob/ac-1.25.1/pkg/storage/storage.go#L140-L144
			Labels: map[string]string{
				"owner":                 owner,
				labelScopeReleaseName:   rls.Name,
				labelScopeReleaseStatus: release.StatusDeployed.String(),
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
				Name:    rls.Name,
				Version: strconv.Itoa(rls.Version),
				Status:  release.StatusDeployed.String(),
				// Status:        string(rls.Info.Status),
				FirstDeployed: &metav1.Time{Time: rls.Info.FirstDeployed.Time.UTC()},
				LastDeployed:  &metav1.Time{Time: rls.Info.LastDeployed.Time.UTC()},
			},
			Components:   nil,
			Selector:     nil,
			ResourceKeys: nil,
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

	if data, ok := rls.Chart.Metadata.Annotations["meta.x-helm.dev/editor"]; ok && data != "" {
		var gvr metav1.GroupVersionResource
		if err := json.Unmarshal([]byte(data), &gvr); err != nil {
			return nil, err
		} else {
			obj.Spec.Editor = &gvr
		}
	}

	//lbl := map[string]string{
	//	"app.kubernetes.io/managed-by": "Helm",
	//}
	lbl := map[string]string{}
	if partOf, ok := rls.Chart.Metadata.Annotations["app.kubernetes.io/part-of"]; ok && partOf != "" {
		lbl["app.kubernetes.io/part-of"] = partOf
	} else {
		lbl["app.kubernetes.io/instance"] = rls.Name

		// ref : https://github.com/x-helm/helm/blob/ac-1.25.1/pkg/action/validate.go#L211-L217
		if obj.Spec.Editor != nil {
			lbl["app.kubernetes.io/name"] = fmt.Sprintf("%s.%s", obj.Spec.Editor.Resource, obj.Spec.Editor.Group)
		}
	}
	obj.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: lbl,
	}

	if obj.Spec.Editor != nil {
		if f, ok := rls.Config["form"]; ok {
			if fd, err := json.Marshal(f); err == nil {
				obj.Spec.Release.Form = &runtime.RawExtension{Raw: fd}
			}
		}

		obj.Spec.ResourceKeys = strings.Split(rls.Chart.Metadata.Annotations["meta.x-helm.dev/resource-keys"], ",")
		obj.Spec.FormKeys = strings.Split(rls.Chart.Metadata.Annotations["meta.x-helm.dev/form-keys"], ",")
	}

	components := make(map[metav1.GroupVersionKind]struct{})
	if data, ok := rls.Chart.Metadata.Annotations["meta.x-helm.dev/resources"]; ok && data != "" {
		var gvks []metav1.GroupVersionKind
		err := yaml.Unmarshal([]byte(data), &gvks)
		if err != nil {
			return nil, err
		}
		for _, gvk := range gvks {
			components[gvk] = empty
		}
	} else {
		var err error
		components, _, err = parser.ExtractComponentGVKs([]byte(rls.Manifest))
		if err != nil {
			// WARNING(tamal): This error should never happen
			return nil, err
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

	return obj, nil
}

// decodeRelease decodes the bytes of data into a release
// type. Data must contain a base64 encoded gzipped string of a
// valid release, otherwise an error is returned.
func decodeReleaseFromApp(kc client.Client, app *driversapi.AppRelease) (*rspb.Release, error) {
	var rls rspb.Release

	rls.Name = app.Name
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
	// we don't want to load chart from remote here anymore, because we want to embed chart in Go binary

	return &rls, nil
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
	formKeys := sets.NewString(app.Spec.FormKeys...)

	for _, gvk := range app.Spec.Components {
		var list unstructured.UnstructuredList
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind,
		})
		err = kc.List(context.TODO(), &list, client.InNamespace(rls.Namespace), client.MatchingLabelsSelector{Selector: selector})
		if meta.IsNoMatchError(err) {
			continue
		} else if err != nil {
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
				if !resourceKeys.Has(rsKey) && !formKeys.Has(rsKey) {
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

			if resourceKeys.Has(rsKey) {
				if _, ok := resourceMap[rsKey]; ok {
					return nil, fmt.Errorf("duplicate resource key %s for AppRelease %s/%s", rsKey, app.Namespace, app.Name)
				}
				resourceMap[rsKey] = &obj
			}
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

func GenerateAppReleaseObject(chrt *chart.Chart, model releasesapi.Metadata) (*driversapi.AppRelease, error) {
	const owner = "helm"

	appName := model.Release.Name
	if partOf, ok := chrt.Metadata.Annotations["app.kubernetes.io/part-of"]; ok {
		appName = partOf
	}
	now := time.Now()

	// create and return configmap object
	obj := &driversapi.AppRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: model.Release.Namespace,
			// labels are required by https://github.com/x-helm/helm/blob/ac-1.25.1/pkg/storage/storage.go#L140-L144
			Labels: map[string]string{
				"owner":                 owner,
				labelScopeReleaseName:   model.Release.Name,
				labelScopeReleaseStatus: release.StatusDeployed.String(),
			},
		},
		Spec: driversapi.AppReleaseSpec{
			Descriptor: driversapi.Descriptor{
				Type:        chrt.Metadata.Type,
				Version:     chrt.Metadata.AppVersion,
				Description: chrt.Metadata.Description, // rls.Info.Description,
				Owners:      nil,                       // FIX
				Keywords:    chrt.Metadata.Keywords,
				Links: []shared.Link{
					{
						Description: "website",
						URL:         chrt.Metadata.Home,
					},
				},
				Notes: "", // rls.Info.Notes,
			},
			Release: driversapi.ReleaseInfo{
				Name:    model.Release.Name,
				Version: "1", // strconv.Itoa(rls.Version),
				Status:  release.StatusDeployed.String(),
				// Status:        string(rls.Info.Status),
				FirstDeployed: &metav1.Time{Time: now.UTC()},
				LastDeployed:  &metav1.Time{Time: now.UTC()},
			},
			Components:   nil,
			Selector:     nil,
			ResourceKeys: nil,
		},
	}
	if chrt.Metadata.Icon != "" {
		var imgType string
		if resp, err := http.Get(chrt.Metadata.Icon); err == nil {
			if mime, err := mimetype.DetectReader(resp.Body); err == nil {
				imgType = mime.String()
			}
			_ = resp.Body.Close()
		}
		obj.Spec.Descriptor.Icons = []shared.ImageSpec{
			{
				Source: chrt.Metadata.Icon,
				// TotalSize: "",
				Type: imgType,
			},
		}
	}
	for _, maintainer := range chrt.Metadata.Maintainers {
		obj.Spec.Descriptor.Maintainers = append(obj.Spec.Descriptor.Maintainers, shared.ContactData{
			Name:  maintainer.Name,
			URL:   maintainer.URL,
			Email: maintainer.Email,
		})
	}

	if data, ok := chrt.Metadata.Annotations["meta.x-helm.dev/editor"]; ok && data != "" {
		var gvr metav1.GroupVersionResource
		if err := json.Unmarshal([]byte(data), &gvr); err != nil {
			return nil, err
		} else {
			obj.Spec.Editor = &gvr
		}
	}

	//lbl := map[string]string{
	//	"app.kubernetes.io/managed-by": "Helm",
	//}
	lbl := map[string]string{}
	if partOf, ok := chrt.Metadata.Annotations["app.kubernetes.io/part-of"]; ok && partOf != "" {
		lbl["app.kubernetes.io/part-of"] = partOf
	} else {
		lbl["app.kubernetes.io/instance"] = model.Release.Name

		// ref : https://github.com/x-helm/helm/blob/ac-1.25.1/pkg/action/validate.go#L211-L217
		if obj.Spec.Editor != nil {
			lbl["app.kubernetes.io/name"] = fmt.Sprintf("%s.%s", obj.Spec.Editor.Resource, obj.Spec.Editor.Group)
		}
	}
	obj.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: lbl,
	}

	if obj.Spec.Editor != nil {
		if f, ok := chrt.Values["form"]; ok {
			if fd, err := json.Marshal(f); err == nil {
				obj.Spec.Release.Form = &runtime.RawExtension{Raw: fd}
			}
		}

		obj.Spec.ResourceKeys = strings.Split(chrt.Metadata.Annotations["meta.x-helm.dev/resource-keys"], ",")
		obj.Spec.FormKeys = strings.Split(chrt.Metadata.Annotations["meta.x-helm.dev/form-keys"], ",")
	}

	if data, ok := chrt.Metadata.Annotations["meta.x-helm.dev/resources"]; ok && data != "" {
		var gvks []metav1.GroupVersionKind
		err := yaml.Unmarshal([]byte(data), &gvks)
		if err != nil {
			return nil, err
		}
		obj.Spec.Components = gvks
	}

	return obj, nil
}

func EnsureAppReleaseCRD(restcfg *rest.Config, mapper meta.RESTMapper) error {
	rsmapper := disco_util.NewResourceMapper(mapper)
	appcrdRegistered, err := rsmapper.ExistsGVR(driversapi.GroupVersion.WithResource("appreleases"))
	if err != nil {
		return fmt.Errorf("failed to detect if AppRelease CRD is registered, reason %v", err)
	}
	if !appcrdRegistered {
		// register AppRelease CRD
		crds := []*apiextensions.CustomResourceDefinition{
			driversapi.AppRelease{}.CustomResourceDefinition(),
		}
		crdClient, err := crd_cs.NewForConfig(restcfg)
		if err != nil {
			return fmt.Errorf("failed to create crd client, reason %v", err)
		}
		err = apiextensions.RegisterCRDs(crdClient, crds)
		if err != nil {
			return fmt.Errorf("failed to register appRelease crd, reason %v", err)
		}
	}
	return nil
}
