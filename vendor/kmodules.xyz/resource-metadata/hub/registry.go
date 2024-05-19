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
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"time"

	"kmodules.xyz/apiversion"
	kmapi "kmodules.xyz/client-go/api/v1"
	disco_util "kmodules.xyz/client-go/discovery"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	"kmodules.xyz/resource-metadata/hub/resourcedescriptors"

	stringz "gomodules.xyz/x/strings"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	CRDIconSVG = "https://cdn.appscode.com/k8s/icons/apiextensions.k8s.io/customresourcedefinitions.svg"
)

// ttl for cached preferred list
const ttl = 5 * time.Minute

type Registry struct {
	uid   string
	cache KV
	m     sync.RWMutex
	// TODO: store in KV so cached for multiple instances of BB api server
	cfg           *rest.Config
	preferred     map[schema.GroupResource]schema.GroupVersionResource
	lastRefreshed time.Time
	regGVK        map[schema.GroupVersionKind]*kmapi.ResourceID
	regGVR        map[schema.GroupVersionResource]*kmapi.ResourceID
}

var _ disco_util.ResourceMapper = &Registry{}

func NewRegistry(uid string, cache KV) *Registry {
	r := &Registry{
		uid:    uid,
		cache:  cache,
		regGVK: map[schema.GroupVersionKind]*kmapi.ResourceID{},
		regGVR: map[schema.GroupVersionResource]*kmapi.ResourceID{},
	}

	guess := make(map[schema.GroupResource]string)

	r.cache.Visit(func(key string, val *v1alpha1.ResourceDescriptor) {
		v := val.Spec.Resource // copy
		r.regGVK[v.GroupVersionKind()] = &v
		r.regGVR[v.GroupVersionResource()] = &v

		gr := v.GroupResource()
		if curVer, ok := guess[gr]; !ok || apiversion.MustCompare(v.Version, curVer) > 0 {
			guess[gr] = v.Version
		}
	})

	r.preferred = make(map[schema.GroupResource]schema.GroupVersionResource)
	for gr, ver := range guess {
		r.preferred[gr] = gr.WithVersion(ver)
	}

	return r
}

func NewRegistryOfKnownResources() *Registry {
	return NewRegistry(KnownUID, &KVMap{
		cache: resourcedescriptors.KnownDescriptors(),
	})
}

func (r *Registry) DiscoverResources(cfg *rest.Config) error {
	r.m.Lock()
	defer r.m.Unlock()

	r.cfg = cfg
	return r.discoverResources()
}

func (r *Registry) discoverResources() error {
	preferred, reg, err := r.createRegistry(r.cfg)
	if err != nil {
		return err
	}

	r.preferred = preferred
	r.lastRefreshed = time.Now()
	for filename, rd := range reg {
		if _, found := r.cache.Get(filename); !found {
			r.regGVK[rd.Spec.Resource.GroupVersionKind()] = &rd.Spec.Resource
			r.regGVR[rd.Spec.Resource.GroupVersionResource()] = &rd.Spec.Resource
			r.cache.Set(filename, rd)
		}
	}

	return nil
}

func (r *Registry) Refresh(cfg *rest.Config) error {
	if time.Since(r.lastRefreshed) > ttl {
		return r.DiscoverResources(cfg)
	}
	return nil
}

func (r *Registry) Reset() {
	r.m.Lock()
	defer r.m.Unlock()
	if r.cfg == nil {
		return
	}
	if err := r.discoverResources(); err != nil {
		klog.ErrorS(err, "failed to reset Registry")
	}
}

func (r *Registry) Register(gvr schema.GroupVersionResource, cfg *rest.Config) error {
	r.m.RLock()
	if _, found := r.regGVR[gvr]; found {
		r.m.RUnlock()
		return nil
	}
	r.m.RUnlock()

	return r.DiscoverResources(cfg)
}

func (r *Registry) createRegistry(cfg *rest.Config) (map[schema.GroupResource]schema.GroupVersionResource, map[string]*v1alpha1.ResourceDescriptor, error) {
	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	apiext, err := crd_cs.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	rsLists, err := kc.Discovery().ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, nil, err
	}

	reg := make(map[string]*v1alpha1.ResourceDescriptor)
	for _, rsList := range rsLists {
		for i := range rsList.APIResources {
			rs := rsList.APIResources[i]

			// skip sub resource
			if strings.ContainsRune(rs.Name, '/') {
				continue
			}
			// if resource can't be listed or read (get) skip it
			if !stringz.Contains(rs.Verbs, "list") || !stringz.Contains(rs.Verbs, "get") {
				continue
			}

			gv, err := schema.ParseGroupVersion(rsList.GroupVersion)
			if err != nil {
				return nil, nil, err
			}
			scope := kmapi.ClusterScoped
			if rs.Namespaced {
				scope = kmapi.NamespaceScoped
			}
			rid := kmapi.ResourceID{
				Group:   gv.Group,
				Version: gv.Version,
				Name:    rs.Name,
				Kind:    rs.Kind,
				Scope:   scope,
			}

			name := resourcedescriptors.GetName(rid.GroupVersionResource())
			rd := v1alpha1.ResourceDescriptor{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       v1alpha1.ResourceKindResourceDescriptor,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"k8s.io/group":    rid.Group,
						"k8s.io/version":  rid.Version,
						"k8s.io/resource": rid.Name,
						"k8s.io/kind":     rid.Kind,
					},
				},
				Spec: v1alpha1.ResourceDescriptorSpec{
					Resource: rid,
				},
			}
			if !meta_util.IsOfficialType(rd.Spec.Resource.Group) {
				crd, err := apiext.CustomResourceDefinitions().Get(context.TODO(), fmt.Sprintf("%s.%s", rd.Spec.Resource.Name, rd.Spec.Resource.Group), metav1.GetOptions{})
				if err == nil {
					for _, v := range crd.Spec.Versions {
						if v.Name == rid.Version {
							rd.Spec.Validation = v.Schema
							break
						}
					}
				}
			}
			reg[name] = &rd
		}
	}

	preferred := make(map[schema.GroupResource]schema.GroupVersionResource)
	for _, rd := range reg {
		gvr := rd.Spec.Resource.GroupVersionResource()
		preferred[gvr.GroupResource()] = gvr
	}

	err = fs.WalkDir(resourcedescriptors.EmbeddedFS(), ".", func(filename string, e fs.DirEntry, err error) error {
		if !e.IsDir() {
			delete(reg, filename)
		}
		return err
	})
	return preferred, reg, err
}

func (r *Registry) Visit(f func(key string, val *v1alpha1.ResourceDescriptor)) {
	for _, gvr := range r.Resources() {
		key := resourcedescriptors.GetName(gvr)
		if rd, ok := r.cache.Get(key); ok {
			f(key, rd)
		}
	}
}

func (r *Registry) Missing(in schema.GroupVersionResource) bool {
	r.m.RLock()
	defer r.m.RUnlock()
	for _, gvr := range r.preferred {
		if gvr == in {
			return false
		}
	}
	return true
}

func (r *Registry) FindGVR(in *metav1.GroupKind, keepOfficialTypes bool) (schema.GroupVersionResource, bool) {
	r.m.RLock()
	defer r.m.RUnlock()

	latestGVRs := resourcedescriptors.LatestGVRs()

	for gvk, rid := range r.regGVK {
		if gvk.Group == in.Group && gvk.Kind == in.Kind {
			if gvr, ok := r.preferred[rid.GroupResource()]; ok {
				return gvr, true
			}
		}
	}

	gk := schema.GroupKind{Group: in.Group, Kind: in.Kind}
	if keepOfficialTypes || !meta_util.IsOfficialType(in.Group) {
		gvr, ok := latestGVRs[gk]
		return gvr, ok
	}
	return schema.GroupVersionResource{}, false
}

func (r *Registry) ResourceIDForGVK(in schema.GroupVersionKind) (*kmapi.ResourceID, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if in.Version != "" {
		return r.regGVK[in], nil
	}
	for gvk, rid := range r.regGVK {
		if gvk.Group == in.Group && gvk.Kind == in.Kind {
			if _, ok := r.preferred[rid.GroupResource()]; ok {
				return rid, nil
			}
		}
	}
	return nil, nil
}

func (r *Registry) ResourceIDForGVR(in schema.GroupVersionResource) (*kmapi.ResourceID, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if in.Version != "" {
		return r.regGVR[in], nil
	}
	for gvr, rid := range r.regGVR {
		if gvr.Group == in.Group && gvr.Resource == in.Resource {
			if _, ok := r.preferred[rid.GroupResource()]; ok {
				return rid, nil
			}
		}
	}
	return nil, nil
}

func (r *Registry) GVR(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exists := r.regGVK[gvk]
	if !exists {
		return schema.GroupVersionResource{}, UnregisteredErr{gvk.String()}
	}
	return rid.GroupVersionResource(), nil
}

func (r *Registry) TypeMeta(gvr schema.GroupVersionResource) (metav1.TypeMeta, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exists := r.regGVR[gvr]
	if !exists {
		return metav1.TypeMeta{}, UnregisteredErr{gvr.String()}
	}
	return rid.TypeMeta(), nil
}

func (r *Registry) GVK(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exists := r.regGVR[gvr]
	if !exists {
		return schema.GroupVersionKind{}, UnregisteredErr{gvr.String()}
	}
	return rid.GroupVersionKind(), nil
}

func (r *Registry) IsGVRNamespaced(gvr schema.GroupVersionResource) (bool, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exist := r.regGVR[gvr]
	if !exist {
		return false, UnregisteredErr{gvr.String()}
	}
	return rid.Scope == kmapi.NamespaceScoped, nil
}

func (r *Registry) IsGVKNamespaced(gvr schema.GroupVersionKind) (bool, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exist := r.regGVK[gvr]
	if !exist {
		return false, UnregisteredErr{gvr.String()}
	}
	return rid.Scope == kmapi.NamespaceScoped, nil
}

func (r *Registry) IsPreferred(gvr schema.GroupVersionResource) (bool, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	if preferred, exists := r.preferred[gvr.GroupResource()]; exists {
		return preferred == gvr, nil
	}
	return false, nil
}

func (r *Registry) Preferred(gvr schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	if preferred, exists := r.preferred[gvr.GroupResource()]; exists {
		return preferred, nil
	}
	return gvr, nil
}

func (r *Registry) ExistsGVR(gvr schema.GroupVersionResource) (bool, error) {
	if gvr.Version == "" {
		return false, fmt.Errorf("unspecificed version for %v", gvr.GroupResource())
	}
	r.m.RLock()
	defer r.m.RUnlock()
	_, exists := r.regGVR[gvr]
	return exists, nil
}

func (r *Registry) ExistsGVK(gvk schema.GroupVersionKind) (bool, error) {
	if gvk.Version == "" {
		return false, fmt.Errorf("unspecificed version for %v", gvk.GroupKind())
	}
	r.m.RLock()
	defer r.m.RUnlock()
	_, exists := r.regGVK[gvk]
	return exists, nil
}

func (r *Registry) Resources() []schema.GroupVersionResource {
	r.m.RLock()
	defer r.m.RUnlock()

	out := make([]schema.GroupVersionResource, len(r.preferred))
	for _, gvr := range r.preferred {
		out = append(out, gvr)
	}
	return out
}

func (r *Registry) Kinds() []schema.GroupVersionKind {
	r.m.RLock()
	defer r.m.RUnlock()

	out := make([]schema.GroupVersionKind, len(r.preferred))
	for _, gvr := range r.preferred {
		out = append(out, r.regGVR[gvr].GroupVersionKind())
	}
	return out
}

func (r *Registry) LoadByGVR(gvr schema.GroupVersionResource) (*v1alpha1.ResourceDescriptor, error) {
	return r.LoadByFile(resourcedescriptors.GetName(gvr))
}

func (r *Registry) LoadByGVK(gvk schema.GroupVersionKind) (*v1alpha1.ResourceDescriptor, error) {
	gvr, err := r.GVR(gvk)
	if err != nil {
		return nil, err
	}
	return r.LoadByFile(resourcedescriptors.GetName(gvr))
}

func (r *Registry) LoadByName(name string) (*v1alpha1.ResourceDescriptor, error) {
	return r.LoadByFile(toFilename(name))
}

func toFilename(name string) string {
	name = reverse(name)
	name = strings.Replace(name, "-", "/", 2)
	name = reverse(name)
	filename := name + ".yaml"
	return filename
}

func (r *Registry) LoadByFile(filename string) (*v1alpha1.ResourceDescriptor, error) {
	obj, ok := r.cache.Get(filename)
	if !ok {
		return nil, UnregisteredErr{filename}
	}
	return obj, nil
}

type UnregisteredErr struct {
	t string
}

var _ error = UnregisteredErr{}

func (e UnregisteredErr) Error() string {
	return e.t + " isn't registered"
}

func IsUnregisteredErr(err error) bool {
	_, ok := err.(UnregisteredErr)
	_, okp := err.(*UnregisteredErr)
	return err != nil && (ok || okp)
}

func ParseGVR(name string) (*schema.GroupVersionResource, error) {
	name = reverse(name)
	parts := strings.SplitN(name, "-", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("%s is not a valid gvr encoded name", name)
	}
	gvr := schema.GroupVersionResource{
		Group:    reverse(parts[2]),
		Version:  reverse(parts[1]),
		Resource: reverse(parts[0]),
	}
	if gvr.Group == "core" {
		gvr.Group = ""
	}
	return &gvr, nil
}

// ref: https://groups.google.com/g/golang-nuts/c/oPuBaYJ17t4/m/PCmhdAyrNVkJ
func reverse(input string) string {
	// Get Unicode code points.
	n := 0
	rune := make([]int32, len(input))
	for _, r := range input {
		rune[n] = r
		n++
	}
	rune = rune[0:n]

	// Reverse
	for i := 0; i < n/2; i++ {
		rune[i], rune[n-1-i] = rune[n-1-i], rune[i]
	}

	// Convert back to UTF-8.
	return string(rune)
}
