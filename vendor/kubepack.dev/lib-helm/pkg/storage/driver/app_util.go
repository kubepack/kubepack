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
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	rspb "helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
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
				"owner":                                  owner,
				"name.release.x-helm.dev/" + rls.Name:    rls.Name,
				"status.release.x-helm.dev/" + rls.Name:  release.StatusDeployed.String(),
				"version.release.x-helm.dev/" + rls.Name: strconv.Itoa(rls.Version),
			},
			Annotations: map[string]string{
				"first-deployed.release.x-helm.dev/" + rls.Name: rls.Info.FirstDeployed.UTC().Format(time.RFC3339),
				"last-deployed.release.x-helm.dev/" + rls.Name:  rls.Info.LastDeployed.UTC().Format(time.RFC3339),
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
			ComponentGroupKinds: nil,
			Selector:            nil,
			AddOwnerRef:         false, // TODO
			AssemblyPhase:       toAssemblyPhase(rls.Info.Status),
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
		obj.Annotations["editor.x-helm.dev/"+rls.Name] = editorGVR

		if f, ok := rls.Config["form"]; ok {
			if fd, err := json.Marshal(f); err == nil {
				obj.Annotations["form.release.x-helm.dev/"+rls.Name] = string(fd)
			}
		}
	}

	components, _, err := parser.ExtractComponentGKs([]byte(rls.Manifest))
	if err != nil {
		// WARNING(tamal): This error should never happen
		panic(err)
	}

	if data, ok := rls.Chart.Metadata.Annotations["meta.x-helm.dev/resources"]; ok && data != "" {
		var gks []metav1.GroupKind
		err := yaml.Unmarshal([]byte(data), &gks)
		if err != nil {
			panic(err)
		}
		for _, gk := range gks {
			components[gk] = empty
		}
	}
	gks := make([]metav1.GroupKind, 0, len(components))
	for gk := range components {
		gks = append(gks, gk)
	}
	sort.Slice(gks, func(i, j int) bool {
		if gks[i].Group == gks[j].Group {
			return gks[i].Kind < gks[j].Kind
		}
		return gks[i].Group < gks[j].Group
	})
	obj.Spec.ComponentGroupKinds = gks

	return obj
}

func toAssemblyPhase(status release.Status) driversapi.AppReleaseAssemblyPhase {
	switch status {
	case release.StatusUnknown, release.StatusUninstalling, release.StatusPendingInstall, release.StatusPendingUpgrade, release.StatusPendingRollback:
		return driversapi.Pending
	case release.StatusDeployed, release.StatusUninstalled, release.StatusSuperseded:
		return driversapi.Succeeded
	case release.StatusFailed:
		return driversapi.Failed
	}
	panic(fmt.Sprintf("unknown status: %s", status.String()))
}

// decodeRelease decodes the bytes of data into a release
// type. Data must contain a base64 encoded gzipped string of a
// valid release, otherwise an error is returned.
func decodeReleaseFromApp(kc client.Client, app *driversapi.AppRelease, rlsNames []string) ([]*rspb.Release, error) {
	if len(rlsNames) == 0 {
		rlsNames = relevantReleases(app.Labels)
	}

	releases := make([]*rspb.Release, 0, len(rlsNames))

	for _, rlsName := range rlsNames {
		var rls rspb.Release

		rls.Name = app.Labels["name.release.x-helm.dev/"+rlsName]
		rls.Namespace = app.Namespace
		rls.Version, _ = strconv.Atoi(app.Labels["version.release.x-helm.dev/"+rlsName])

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
			Status:      release.Status(app.Labels["status.release.x-helm.dev/"+rlsName]),
			Notes:       app.Spec.Descriptor.Notes,
		}
		rls.Info.FirstDeployed, _ = helmtime.Parse(time.RFC3339, app.Annotations["first-deployed.release.x-helm.dev/"+rlsName])
		rls.Info.LastDeployed, _ = helmtime.Parse(time.RFC3339, app.Annotations["last-deployed.release.x-helm.dev/"+rlsName])

		var editorGVR *metav1.GroupVersionResource
		if data, ok := app.Annotations["editor.x-helm.dev/"+rlsName]; ok && data != "" {
			var gvr metav1.GroupVersionResource
			err := yaml.Unmarshal([]byte(data), gvr)
			if err != nil {
				return nil, fmt.Errorf("editor.x-helm.dev/%s is not a valid GVR, reason %v", rlsName, err)
			}
			editorGVR = &gvr
		}
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
			if f, ok := app.Annotations["form.release.x-helm.dev/"+rls.Name]; ok {
				var form map[string]interface{}
				if err = json.Unmarshal([]byte(f), &form); err == nil {
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

	for _, gk := range app.Spec.ComponentGroupKinds {
		mapping, err := kc.RESTMapper().RESTMapping(schema.GroupKind{
			Group: gk.Group,
			Kind:  gk.Kind,
			// Version: "", // use the preferred version
		})
		if err != nil {
			klog.Warningf("failed to detect GVR for gk %v, reason %v", gk, err)
			continue
		}
		gvr := mapping.Resource

		var list unstructured.UnstructuredList
		list.SetGroupVersionKind(mapping.GroupVersionKind)
		err = kc.List(context.TODO(), &list, client.InNamespace(rls.Namespace), client.MatchingLabelsSelector{Selector: selector})
		if err != nil {
			return nil, err
		}
		for _, obj := range list.Items {
			// check ownership
			// https://github.com/kubepack/helm/blob/ac-1.21.0/pkg/action/validate.go#L87-L92

			annotationMap := obj.GetAnnotations()
			if v := annotationMap["meta.helm.sh/release-name"]; v != rls.Name {
				continue
			}
			if v := annotationMap["meta.helm.sh/release-namespace"]; v != rls.Namespace {
				continue
			}

			// check that we got the right version
			if v, ok := annotationMap[core.LastAppliedConfigAnnotation]; !ok {
				return nil, fmt.Errorf("failed to detect version for GK %#v in release %v", gk, rls)
			} else {
				var mt metav1.TypeMeta
				err := json.Unmarshal([]byte(v), &mt)
				if err != nil {
					return nil, fmt.Errorf("failed to parse TypeMeta from %s", v)
				}
				gv, err := schema.ParseGroupVersion(mt.APIVersion)
				if err != nil {
					return nil, fmt.Errorf("failed to parse version from %v for %s", mt.APIVersion, v)
				}
				if gv.Version != gvr.Version {
					// object not using preferred version, so we need to load the correct version again

					var o2 unstructured.Unstructured
					o2.SetGroupVersionKind(mapping.GroupVersionKind)
					o2.SetAPIVersion(gv.Version)
					err = kc.Get(context.TODO(), client.ObjectKeyFromObject(&obj), &o2)
					if err != nil {
						return nil, fmt.Errorf("failed to get object with correct apiVersion, reason %v", err)
					}
					obj = o2
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

			rsKey, err := ResourceKey(obj.GetAPIVersion(), obj.GetKind(), rls.Name, obj.GetName())
			if err != nil {
				return nil, err
			}
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
