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
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	rspb "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kblabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/tools/cache"
	cu "kmodules.xyz/client-go/client"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	driversapi "x-helm.dev/apimachinery/apis/drivers/v1alpha1"
)

const (
	labelScopeReleaseName   = "release.x-helm.dev/name"
	labelScopeReleaseStatus = "release.x-helm.dev/status"
)

var _ driver.Driver = (*AppReleases)(nil)

// AppReleasesDriverName is the string name of the driver.
const AppReleasesDriverName = "drivers.x-helm.dev/appreleases"

// AppReleases is a wrapper around an implementation of a kubernetes
// AppReleasesInterface.
type AppReleases struct {
	kc  client.Client
	Log func(string, ...interface{})
}

// NewAppReleases initializes a new AppReleases wrapping an implementation of
// the kubernetes AppReleasesInterface.
func NewAppReleases(ai client.Client) *AppReleases {
	return &AppReleases{
		kc:  ai,
		Log: func(_ string, _ ...interface{}) {},
	}
}

// Name returns the name of the driver.
func (d *AppReleases) Name() string {
	return AppReleasesDriverName
}

// Get fetches the release named by key. The corresponding release is returned
// or error if not found.
func (d *AppReleases) Get(key string) (*rspb.Release, error) {
	relName, _, err := ParseKey(key)
	if err != nil {
		return nil, err
	}

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			labelScopeReleaseName: relName,
		},
	})
	if err != nil {
		return nil, err
	}

	// fetch the configmap holding the release named by key
	var result driversapi.AppReleaseList
	err = d.kc.List(context.Background(), &result, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		d.Log("get: failed to get release %q: %s", relName, err)
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, driver.ErrReleaseNotFound
	}
	if len(result.Items) > 1 {
		names := make([]string, 0, len(result.Items))
		for _, obj := range result.Items {
			name, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return nil, err
			}
			names = append(names, name)
		}
		return nil, fmt.Errorf("multiple matching appRelease objects found %s", strings.Join(names, ","))
	}
	obj := &result.Items[0]

	// found the configmap, decode the base64 data string
	rls, err := decodeReleaseFromApp(d.kc, obj)
	if err != nil {
		d.Log("get: failed to decode data %q: %s", key, err)
		return nil, err
	}
	// return the release object
	return rls, nil
}

// List fetches all releases and returns the list releases such
// that filter(release) == true. An error is returned if the
// configmap fails to retrieve the releases.
func (d *AppReleases) List(filter func(*rspb.Release) bool) ([]*rspb.Release, error) {
	var list driversapi.AppReleaseList
	err := d.kc.List(context.Background(), &list, client.MatchingLabels{
		"owner": "helm",
	})
	if err != nil {
		d.Log("list: failed to list: %s", err)
		return nil, err
	}

	var results []*rspb.Release

	// iterate over the configmaps object list
	// and decode each release
	for _, item := range list.Items {
		rls, err := decodeReleaseFromApp(d.kc, &item)
		if err != nil {
			d.Log("list: failed to decode release: %v: %s", item, err)
			continue
		}
		if filter(rls) {
			results = append(results, rls)
		}
	}
	return results, nil
}

// Query fetches all releases that match the provided map of labels.
// An error is returned if the configmap fails to retrieve the releases.
func (d *AppReleases) Query(labels map[string]string) ([]*rspb.Release, error) {
	ls := kblabels.Set{}
	for k, v := range labels {
		if errs := validation.IsValidLabelValue(v); len(errs) != 0 {
			return nil, errors.Errorf("invalid label value: %q: %s", v, strings.Join(errs, "; "))
		}
		ls[k] = v
	}

	var list driversapi.AppReleaseList
	err := d.kc.List(context.Background(), &list, client.MatchingLabels(ls))
	if err != nil {
		d.Log("query: failed to query with labels: %s", err)
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, driver.ErrReleaseNotFound
	}

	var results []*rspb.Release
	for _, item := range list.Items {
		rls, err := decodeReleaseFromApp(d.kc, &item)
		if err != nil {
			d.Log("query: failed to decode release: %s", err)
			continue
		}
		results = append(results, rls)
	}
	return results, nil
}

// Create creates a new AppRelease holding the release. If the
// AppRelease already exists, ErrReleaseExists is returned.
func (d *AppReleases) Create(_ string, rls *rspb.Release) error {
	// create a new configmap to hold the release
	obj := mustNewAppReleaseObject(rls)

	// push the configmap object out into the kubiverse
	_, err := cu.CreateOrPatch(context.Background(), d.kc, obj, func(o client.Object, createOp bool) client.Object {
		in := o.(*driversapi.AppRelease)

		in.Labels = meta_util.OverwriteKeys(in.Labels, obj.Labels)
		in.Annotations = meta_util.OverwriteKeys(in.Annotations, obj.Annotations)

		// merge GKs
		gvkMap := map[metav1.GroupVersionKind]interface{}{}
		for _, gvk := range in.Spec.Components {
			gvkMap[gvk] = empty
		}
		for _, gvk := range obj.Spec.Components {
			gvkMap[gvk] = empty
		}
		gvks := make([]metav1.GroupVersionKind, 0, len(gvkMap))
		for gk := range gvkMap {
			gvks = append(gvks, gk)
		}
		sort.Slice(gvks, func(i, j int) bool {
			if gvks[i].Group == gvks[j].Group {
				return gvks[i].Kind < gvks[j].Kind
			}
			return gvks[i].Group < gvks[j].Group
		})

		if err := mergo.Merge(&in.Spec, &obj.Spec); err != nil {
			panic(fmt.Errorf("failed to update appliation %s/%s spec, reason: %v", in.Namespace, in.Name, err))
		}
		in.Spec.Selector = obj.Spec.Selector
		in.Spec.Components = gvks
		in.Spec.Release = obj.Spec.Release
		return in
	})
	if err != nil {
		//if apierrors.IsAlreadyExists(err) {
		//	return driver.ErrReleaseExists
		//}

		d.Log("create: failed to create: %s", err)
		return err
	}
	return nil
}

// Update updates the AppRelease holding the release. If not found
// the AppRelease is created to hold the release.
func (d *AppReleases) Update(_ string, rls *rspb.Release) error {
	// Bypass update call if called on originalRelease.
	// Update() just updates the modifiedAt timestamp. This is not that important for our app driver.
	if rls.Chart == nil || rls.Chart.Metadata == nil {
		return nil
	}

	// create a new configmap object to hold the release
	obj := mustNewAppReleaseObject(rls)
	obj.Spec.Release.ModifiedAt = &metav1.Time{Time: time.Now().UTC()}

	// push the configmap object out into the kubiverse
	_, err := cu.CreateOrPatch(context.Background(), d.kc, obj, func(o client.Object, createOp bool) client.Object {
		in := o.(*driversapi.AppRelease)

		in.Labels = meta_util.MergeKeys(in.Labels, obj.Labels)
		in.Annotations = meta_util.MergeKeys(in.Annotations, obj.Annotations)
		if err := mergo.Merge(&in.Spec, &obj.Spec); err != nil {
			panic(fmt.Errorf("failed to update appliation %s/%s spec, reason: %v", in.Namespace, in.Name, err))
		}
		in.Spec.Release = obj.Spec.Release
		return in
	})
	if err != nil {
		d.Log("update: failed to update: %s", err)
		return err
	}
	return nil
}

// Delete deletes the AppRelease holding the release named by key.
func (d *AppReleases) Delete(key string) (rls *rspb.Release, err error) {
	relName, _, err := ParseKey(key)
	if err != nil {
		return nil, err
	}

	// fetch the release to check existence
	if rls, err = d.Get(key); err != nil {
		if err == driver.ErrReleaseNotFound {
			return rls, nil
		}
		return nil, err
	}
	// delete the release
	if err != nil {
		return nil, err
	}
	if err = d.kc.DeleteAllOf(context.Background(), new(driversapi.AppRelease), client.MatchingLabels{
		labelScopeReleaseName: relName,
	}); err != nil {
		return rls, err
	}
	return rls, nil
}
