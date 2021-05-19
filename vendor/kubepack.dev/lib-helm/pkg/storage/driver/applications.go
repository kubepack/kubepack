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

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	rspb "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kblabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/discovery"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/application/api/app/v1beta1"
	cs "sigs.k8s.io/application/client/clientset/versioned/typed/app/v1beta1"
)

const (
	annotaionScopeReleaseName = "name.release.x-helm.dev" // "/${name}" : ""
)

var _ driver.Driver = (*Applications)(nil)

// ApplicationsDriverName is the string name of the driver.
const ApplicationsDriverName = "storage.x-helm.dev/apps"

// Applications is a wrapper around an implementation of a kubernetes
// ApplicationsInterface.
type Applications struct {
	ai  cs.ApplicationInterface
	di  dynamic.Interface
	cl  discovery.ResourceMapper
	Log func(string, ...interface{})
}

// NewApplications initializes a new Applications wrapping an implementation of
// the kubernetes ApplicationsInterface.
func NewApplications(ai cs.ApplicationInterface, di dynamic.Interface, cl discovery.ResourceMapper) *Applications {
	return &Applications{
		ai:  ai,
		di:  di,
		cl:  cl,
		Log: func(_ string, _ ...interface{}) {},
	}
}

// Name returns the name of the driver.
func (d *Applications) Name() string {
	return ApplicationsDriverName
}

// Get fetches the release named by key. The corresponding release is returned
// or error if not found.
func (d *Applications) Get(key string) (*rspb.Release, error) {
	relName, _, err := ParseKey(key)
	if err != nil {
		return nil, err
	}

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			fmt.Sprintf("%s/%s", annotaionScopeReleaseName, relName): relName,
		},
	})
	if err != nil {
		return nil, err
	}

	// fetch the configmap holding the release named by key
	result, err := d.ai.List(context.Background(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
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
		return nil, fmt.Errorf("multiple matching application objects found %s", strings.Join(names, ","))
	}
	obj := &result.Items[0]

	// found the configmap, decode the base64 data string
	r, err := decodeReleaseFromApp(obj, []string{relName}, d.di, d.cl)
	if err != nil {
		d.Log("get: failed to decode data %q: %s", key, err)
		return nil, err
	}
	// return the release object
	return r[0], nil
}

// List fetches all releases and returns the list releases such
// that filter(release) == true. An error is returned if the
// configmap fails to retrieve the releases.
func (d *Applications) List(filter func(*rspb.Release) bool) ([]*rspb.Release, error) {
	lsel := kblabels.Set{"owner": "helm"}.AsSelector()
	opts := metav1.ListOptions{LabelSelector: lsel.String()}

	list, err := d.ai.List(context.Background(), opts)
	if err != nil {
		d.Log("list: failed to list: %s", err)
		return nil, err
	}

	var results []*rspb.Release

	// iterate over the configmaps object list
	// and decode each release
	for _, item := range list.Items {
		releases, err := decodeReleaseFromApp(&item, nil, d.di, d.cl)
		if err != nil {
			d.Log("list: failed to decode release: %v: %s", item, err)
			continue
		}
		for _, rls := range releases {
			if filter(rls) {
				results = append(results, rls)
			}
		}
	}
	return results, nil
}

// Query fetches all releases that match the provided map of labels.
// An error is returned if the configmap fails to retrieve the releases.
func (d *Applications) Query(labels map[string]string) ([]*rspb.Release, error) {
	ls := kblabels.Set{}
	for k, v := range labels {
		if errs := validation.IsValidLabelValue(v); len(errs) != 0 {
			return nil, errors.Errorf("invalid label value: %q: %s", v, strings.Join(errs, "; "))
		}
		ls[k] = v
	}

	opts := metav1.ListOptions{LabelSelector: ls.AsSelector().String()}

	list, err := d.ai.List(context.Background(), opts)
	if err != nil {
		d.Log("query: failed to query with labels: %s", err)
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, driver.ErrReleaseNotFound
	}

	var results []*rspb.Release
	for _, item := range list.Items {
		releases, err := decodeReleaseFromApp(&item, relevantReleases(labels), d.di, d.cl)
		if err != nil {
			d.Log("query: failed to decode release: %s", err)
			continue
		}
		results = append(results, releases...)
	}
	return results, nil
}

// Create creates a new Application holding the release. If the
// Application already exists, ErrReleaseExists is returned.
func (d *Applications) Create(_ string, rls *rspb.Release) error {
	// create a new configmap to hold the release
	obj := newApplicationObject(rls)

	// push the configmap object out into the kubiverse
	_, _, err := createOrPatchApplication(context.Background(), d.ai, obj.ObjectMeta, func(in *v1beta1.Application) *v1beta1.Application {
		in.Labels = meta_util.OverwriteKeys(in.Labels, obj.Labels)
		in.Annotations = meta_util.OverwriteKeys(in.Annotations, obj.Annotations)

		// merge GKs
		gkMap := map[metav1.GroupKind]interface{}{}
		for _, gk := range in.Spec.ComponentGroupKinds {
			gkMap[gk] = empty
		}
		for _, gk := range obj.Spec.ComponentGroupKinds {
			gkMap[gk] = empty
		}
		gks := make([]metav1.GroupKind, 0, len(gkMap))
		for gk := range gkMap {
			gks = append(gks, gk)
		}
		sort.Slice(gks, func(i, j int) bool {
			if gks[i].Group == gks[j].Group {
				return gks[i].Kind < gks[j].Kind
			}
			return gks[i].Group < gks[j].Group
		})

		if err := mergo.Merge(&in.Spec, &obj.Spec); err != nil {
			panic(fmt.Errorf("failed to update appliation %s/%s spec, reason: %v", in.Namespace, in.Name, err))
		}
		in.Spec.Selector = obj.Spec.Selector
		in.Spec.ComponentGroupKinds = gks
		in.Spec.AssemblyPhase = obj.Spec.AssemblyPhase
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		//if apierrors.IsAlreadyExists(err) {
		//	return driver.ErrReleaseExists
		//}

		d.Log("create: failed to create: %s", err)
		return err
	}
	return nil
}

// Update updates the Application holding the release. If not found
// the Application is created to hold the release.
func (d *Applications) Update(_ string, rls *rspb.Release) error {
	// create a new configmap object to hold the release
	obj := newApplicationObject(rls)

	// push the configmap object out into the kubiverse
	_, _, err := createOrPatchApplication(context.Background(), d.ai, obj.ObjectMeta, func(in *v1beta1.Application) *v1beta1.Application {
		in.Labels = meta_util.MergeKeys(in.Labels, obj.Labels)
		in.Annotations = meta_util.MergeKeys(in.Annotations, obj.Annotations)
		if err := mergo.Merge(&in.Spec, &obj.Spec); err != nil {
			panic(fmt.Errorf("failed to update appliation %s/%s spec, reason: %v", in.Namespace, in.Name, err))
		}
		in.Spec.AssemblyPhase = obj.Spec.AssemblyPhase
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		d.Log("update: failed to update: %s", err)
		return err
	}
	return nil
}

// Delete deletes the Application holding the release named by key.
func (d *Applications) Delete(key string) (rls *rspb.Release, err error) {
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
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			fmt.Sprintf("%s/%s", annotaionScopeReleaseName, relName): relName,
		},
	})
	if err != nil {
		return nil, err
	}
	if err = d.ai.DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: selector.String(),
	}); err != nil {
		return rls, err
	}
	return rls, nil
}
