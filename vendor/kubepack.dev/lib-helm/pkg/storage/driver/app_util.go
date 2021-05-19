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
	"k8s.io/client-go/dynamic"
	"kmodules.xyz/client-go/discovery"
	"kmodules.xyz/client-go/tools/parser"
	"sigs.k8s.io/application/api/app/v1beta1"
	"sigs.k8s.io/yaml"
)

var empty = struct{}{}

// newApplicationSecretsObject constructs a kubernetes Application object
// to store a release. Each configmap data entry is the base64
// encoded gzipped string of a release.
//
// The following labels are used within each configmap:
//
//    "modifiedAt"     - timestamp indicating when this configmap was last modified. (set in Update)
//    "createdAt"      - timestamp indicating when this configmap was created. (set in Create)
//    "version"        - version of the release.
//    "status"         - status of the release (see pkg/release/status.go for variants)
//    "owner"          - owner of the configmap, currently "helm".
//    "name"           - name of the release.
//
func newApplicationObject(rls *rspb.Release) *v1beta1.Application {
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
	obj := &v1beta1.Application{
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
		Spec: v1beta1.ApplicationSpec{
			Descriptor: v1beta1.Descriptor{
				Type:        rls.Chart.Metadata.Type,
				Version:     rls.Chart.Metadata.AppVersion,
				Description: rls.Info.Description,
				Owners:      nil, // FIX
				Keywords:    rls.Chart.Metadata.Keywords,
				Links: []v1beta1.Link{
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
		obj.Spec.Descriptor.Icons = []v1beta1.ImageSpec{
			{
				Source: rls.Chart.Metadata.Icon,
				// TotalSize: "",
				Type: imgType,
			},
		}
	}
	for _, maintainer := range rls.Chart.Metadata.Maintainers {
		obj.Spec.Descriptor.Maintainers = append(obj.Spec.Descriptor.Maintainers, v1beta1.ContactData{
			Name:  maintainer.Name,
			URL:   maintainer.URL,
			Email: maintainer.Email,
		})
	}

	lbl := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
	}
	if partOf, ok := rls.Chart.Metadata.Annotations["app.kubernetes.io/part-of"]; ok && partOf != "" {
		lbl["app.kubernetes.io/part-of"] = partOf
	} else {
		lbl["app.kubernetes.io/name"] = rls.Chart.Name()
		lbl["app.kubernetes.io/instance"] = rls.Name
	}
	obj.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: lbl,
	}

	if editorGVR, ok := rls.Chart.Metadata.Annotations["meta.x-helm.dev/editor"]; ok {
		obj.Annotations["editor.x-helm.dev/"+rls.Name] = editorGVR
	}

	components, _, err := parser.ExtractComponents([]byte(rls.Manifest))
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

func toAssemblyPhase(status release.Status) v1beta1.ApplicationAssemblyPhase {
	switch status {
	case release.StatusUnknown, release.StatusUninstalling, release.StatusPendingInstall, release.StatusPendingUpgrade, release.StatusPendingRollback:
		return v1beta1.Pending
	case release.StatusDeployed, release.StatusUninstalled, release.StatusSuperseded:
		return v1beta1.Succeeded
	case release.StatusFailed:
		return v1beta1.Failed
	}
	panic(fmt.Sprintf("unknown status: %s", status.String()))
}

// decodeRelease decodes the bytes of data into a release
// type. Data must contain a base64 encoded gzipped string of a
// valid release, otherwise an error is returned.
func decodeReleaseFromApp(app *v1beta1.Application, rlsNames []string, di dynamic.Interface, cl discovery.ResourceMapper) ([]*rspb.Release, error) {
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
		//	return nil, fmt.Errorf("missing %s annotation on application %s/%s", apis.LabelChartURL, app.Namespace, app.Name)
		//}
		//chartName, ok := app.Annotations[apis.LabelChartName]
		//if !ok {
		//	return nil, fmt.Errorf("missing %s annotation on application %s/%s", apis.LabelChartName, app.Namespace, app.Name)
		//}
		//chartVersion, ok := app.Annotations[apis.LabelChartVersion]
		//if !ok {
		//	return nil, fmt.Errorf("missing %s annotation on application %s/%s", apis.LabelChartVersion, app.Namespace, app.Name)
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
		tpl, err := EditorChartValueManifest(app, cl, di, rlm, editorGVR)
		if err != nil {
			return nil, err
		}

		rls.Manifest = string(tpl.Manifest)

		if editorGVR != nil {
			rls.Chart = &chart.Chart{
				Values: map[string]interface{}{},
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

func EditorChartValueManifest(app *v1beta1.Application, mapper discovery.ResourceMapper, dc dynamic.Interface, rls types.NamespacedName, editorGVR *metav1.GroupVersionResource) (*EditorTemplate, error) {
	labels := app.Spec.Selector
	labels.MatchLabels["app.kubernetes.io/instance"] = rls.Name
	// TODO: keep track of chart name via labels?
	// appNameLabel                   = "app.kubernetes.io/name" not used

	selector, err := metav1.LabelSelectorAsSelector(labels)
	if err != nil {
		return nil, err
	}
	labelSelector := selector.String()

	var buf bytes.Buffer
	resourceMap := map[string]interface{}{}

	for _, gk := range app.Spec.ComponentGroupKinds {
		gvr, err := mapper.GVR(schema.GroupVersionKind{
			Group:   gk.Group,
			Kind:    gk.Kind,
			Version: "", // use the preferred version
		})
		if err != nil {
			return nil, fmt.Errorf("failed to detect GVR for gk %v, reason %v", gk, err)
		}
		namespaced, err := mapper.IsNamespaced(gvr)
		if err != nil {
			return nil, fmt.Errorf("failed to detect if gvr %v is namespaced, reason %v", gvr, err)
		}
		var rc dynamic.ResourceInterface
		if namespaced {
			rc = dc.Resource(gvr).Namespace(rls.Namespace)
		} else {
			rc = dc.Resource(gvr)
		}

		list, err := rc.List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range list.Items {
			// check ownership
			// https://github.com/kubepack/helm/blob/ac-1.21.0/pkg/action/validate.go#L87-L92

			annotationMap := app.GetAnnotations()
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
					o2, err := rc.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
					if err != nil {
						return nil, fmt.Errorf("failed to get object with correct apiVersion, reason %v", err)
					}
					obj = *o2
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
				return nil, fmt.Errorf("duplicate resource key %s for application %s/%s", rsKey, app.Namespace, app.Name)
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
	//groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
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
	//groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
	groupPrefix = strings.Replace(groupPrefix, ".", "_", -1)
	groupPrefix = flect.Pascalize(groupPrefix)

	var nameSuffix string
	nameSuffix = strings.TrimPrefix(name, chartName)
	nameSuffix = strings.Replace(nameSuffix, ".", "-", -1)
	nameSuffix = strings.Trim(nameSuffix, "-")
	nameSuffix = flect.Pascalize(nameSuffix)

	return flect.Underscore(kind), flect.Underscore(kind + nameSuffix), flect.Underscore(groupPrefix + kind + nameSuffix)
}
