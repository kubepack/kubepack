/*
Copyright The Kmodules Authors.

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
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	"kmodules.xyz/resource-metadata/hub/resourceclasses"
	"kmodules.xyz/resource-metadata/hub/resourcedescriptors"

	"gomodules.xyz/version"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Registry struct {
	uid       string
	cache     KV
	m         sync.RWMutex
	preferred []schema.GroupVersionResource
	regGVK    map[schema.GroupVersionKind]*v1alpha1.ResourceID
	regGVR    map[schema.GroupVersionResource]*v1alpha1.ResourceID
}

func NewRegistry(uid string, cache KV) *Registry {
	r := &Registry{
		uid:    uid,
		cache:  cache,
		regGVK: map[schema.GroupVersionKind]*v1alpha1.ResourceID{},
		regGVR: map[schema.GroupVersionResource]*v1alpha1.ResourceID{},
	}

	guess := make(map[schema.GroupResource]string)

	r.cache.Visit(func(key string, val *v1alpha1.ResourceDescriptor) {
		v := val.Spec.Resource // copy
		r.regGVK[v.GroupVersionKind()] = &v
		r.regGVR[v.GroupVersionResource()] = &v

		gr := v.GroupResource()
		if curVer, ok := guess[gr]; !ok || mustCompareVersions(v.Version, curVer) > 0 {
			guess[gr] = v.Version
		}
	})

	r.preferred = make([]schema.GroupVersionResource, 0, len(guess))
	for gr, version := range guess {
		r.preferred = append(r.preferred, gr.WithVersion(version))
	}

	return r
}

func mustCompareVersions(x, y string) int {
	result, err := compareVersions(x, y)
	if err != nil {
		panic(err)
	}
	return result
}

func compareVersions(x, y string) (int, error) {
	xv, err := version.NewVersion(x)
	if err != nil {
		return 0, err
	}
	yv, err := version.NewVersion(y)
	if err != nil {
		return 0, err
	}
	return xv.Compare(yv), nil
}

func NewRegistryOfKnownResources() *Registry {
	return NewRegistry(KnownUID, KnownResources)
}

func (r *Registry) DiscoverResources(cfg *rest.Config) error {
	preferred, reg, err := r.createRegistry(cfg)
	if err != nil {
		return err
	}

	r.m.Lock()
	r.preferred = preferred
	for filename, rd := range reg {
		if _, found := r.cache.Get(filename); !found {
			r.regGVK[rd.Spec.Resource.GroupVersionKind()] = &rd.Spec.Resource
			r.regGVR[rd.Spec.Resource.GroupVersionResource()] = &rd.Spec.Resource
			r.cache.Set(filename, rd)
		}
	}
	r.m.Unlock()

	return nil
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

func (r *Registry) createRegistry(cfg *rest.Config) ([]schema.GroupVersionResource, map[string]*v1alpha1.ResourceDescriptor, error) {
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
			if strings.ContainsRune(rs.Name, '/') {
				continue
			}

			gv, err := schema.ParseGroupVersion(rsList.GroupVersion)
			if err != nil {
				return nil, nil, err
			}
			rs.Group = gv.Group
			rs.Version = gv.Version

			scope := v1alpha1.ClusterScoped
			if rs.Namespaced {
				scope = v1alpha1.NamespaceScoped
			}

			filename := fmt.Sprintf("%s/%s/%s.yaml", rs.Group, rs.Version, rs.Name)
			rd := v1alpha1.ResourceDescriptor{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       v1alpha1.ResourceKindResourceDescriptor,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s-%s", rs.Group, rs.Version, rs.Name),
					Labels: map[string]string{
						"k8s.io/group":    rs.Group,
						"k8s.io/version":  rs.Version,
						"k8s.io/resource": rs.Name,
						"k8s.io/kind":     rs.Kind,
					},
				},
				Spec: v1alpha1.ResourceDescriptorSpec{
					Resource: v1alpha1.ResourceID{
						Group:   rs.Group,
						Version: rs.Version,
						Name:    rs.Name,
						Kind:    rs.Kind,
						Scope:   scope,
					},
				},
			}
			if !rd.Spec.Resource.IsOfficialType() {
				crd, err := apiext.CustomResourceDefinitions().Get(fmt.Sprintf("%s.%s", rd.Spec.Resource.Name, rd.Spec.Resource.Group), metav1.GetOptions{})
				if err == nil {
					rd.Spec.Validation = crd.Spec.Validation
				}
			}
			reg[filename] = &rd
		}
	}

	preferred := make([]schema.GroupVersionResource, 0, len(reg))
	for _, rd := range reg {
		preferred = append(preferred, rd.Spec.Resource.GroupVersionResource())
	}

	for _, name := range resourcedescriptors.AssetNames() {
		delete(reg, name)
	}
	return preferred, reg, nil
}

func (r *Registry) Visit(f func(key string, val *v1alpha1.ResourceDescriptor)) {
	for _, gvr := range r.Resources() {
		key := r.filename(gvr)
		if rd, ok := r.cache.Get(key); ok {
			f(key, rd)
		}
	}
}

func (r *Registry) GVR(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exist := r.regGVK[gvk]
	if !exist {
		return schema.GroupVersionResource{}, UnregisteredErr{gvk.String()}
	}
	return rid.GroupVersionResource(), nil
}

func (r *Registry) TypeMeta(gvr schema.GroupVersionResource) (metav1.TypeMeta, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exist := r.regGVR[gvr]
	if !exist {
		return metav1.TypeMeta{}, UnregisteredErr{gvr.String()}
	}
	return rid.TypeMeta(), nil
}

func (r *Registry) GVK(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exist := r.regGVR[gvr]
	if !exist {
		return schema.GroupVersionKind{}, UnregisteredErr{gvr.String()}
	}
	return rid.GroupVersionKind(), nil
}

func (r *Registry) IsNamespaced(gvr schema.GroupVersionResource) (bool, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	rid, exist := r.regGVR[gvr]
	if !exist {
		return false, UnregisteredErr{gvr.String()}
	}
	return rid.Scope == v1alpha1.NamespaceScoped, nil
}

func (r *Registry) Resources() []schema.GroupVersionResource {
	r.m.RLock()
	defer r.m.RUnlock()

	out := make([]schema.GroupVersionResource, 0, len(r.preferred))
	for _, gvr := range r.preferred {
		out = append(out, gvr)
	}
	return out
}

func (r *Registry) LoadByGVR(gvr schema.GroupVersionResource) (*v1alpha1.ResourceDescriptor, error) {
	return r.LoadByFile(r.filename(gvr))
}

func (r *Registry) filename(gvr schema.GroupVersionResource) string {
	if gvr.Group == "" && gvr.Version == "v1" {
		return fmt.Sprintf("core/v1/%s.yaml", gvr.Resource)
	}
	return fmt.Sprintf("%s/%s/%s.yaml", gvr.Group, gvr.Version, gvr.Resource)
}

func (r *Registry) LoadByName(name string) (*v1alpha1.ResourceDescriptor, error) {
	filename := strings.Replace(name, "-", "/", 2) + ".yaml"
	return r.LoadByFile(filename)
}

func (r *Registry) LoadByFile(filename string) (*v1alpha1.ResourceDescriptor, error) {
	obj, ok := r.cache.Get(filename)
	if !ok {
		return nil, UnregisteredErr{filename}
	}
	return obj, nil
}

func (r *Registry) AvailableResourcePanel() (*v1alpha1.ResourcePanel, error) {
	sections := make(map[string]*v1alpha1.PanelSection)

	// now, auto discover sections from registry
	r.Visit(func(_ string, rd *v1alpha1.ResourceDescriptor) {

		name := resourceclasses.ResourceClassName(rd.Spec.Resource.Group)

		section, found := sections[rd.Spec.Resource.Group]
		if !found {
			if rc, found := KnownClasses[rd.Spec.Resource.Group]; found {
				section = &v1alpha1.PanelSection{
					Name:              rc.Name,
					ResourceClassInfo: rc.Spec.ResourceClassInfo,
				}
			} else {
				// unknown api group, so use CRD icon
				section = &v1alpha1.PanelSection{
					Name: name,
					ResourceClassInfo: v1alpha1.ResourceClassInfo{
						APIGroup: rd.Spec.Resource.Group,
						Icons: []v1alpha1.ImageSpec{
							{
								Source: "https://cdn.appscode.com/k8s/icons/apiextensions.k8s.io/crd.svg",
								Type:   "image/svg+xml",
							},
						},
					},
				}
			}
			sections[rd.Spec.Resource.Group] = section
		}

		section.Entries = append(section.Entries, v1alpha1.PanelEntry{
			Entry: v1alpha1.Entry{
				Name: rd.Spec.Resource.Kind,
				Type: &v1alpha1.GroupVersionResource{
					Group:    rd.Spec.Resource.Group,
					Version:  rd.Spec.Resource.Version,
					Resource: rd.Spec.Resource.Name,
				},
				Icons: rd.Spec.Icons,
			},
			Namespaced: rd.Spec.Resource.Scope == v1alpha1.NamespaceScoped,
		})
	})

	return toPanel(sections)
}

func (r *Registry) DefaultResourcePanel(cfg *rest.Config) (*v1alpha1.ResourcePanel, error) {
	sections := make(map[string]*v1alpha1.PanelSection)
	existingGVRs := map[schema.GroupVersionResource]bool{}

	// first add the known required sections
	for group, rc := range KnownClasses {
		if !rc.IsRequired() {
			continue
		}

		section := &v1alpha1.PanelSection{
			Name:              rc.Name,
			ResourceClassInfo: rc.Spec.ResourceClassInfo,
			Weight:            rc.Spec.Weight,
		}
		for _, entry := range rc.Spec.Entries {
			pe := v1alpha1.PanelEntry{
				Entry:      entry,
				Namespaced: false,
			}
			if entry.Type != nil {
				gvr := entry.Type.GVR()
				existingGVRs[gvr] = true
				if rd, err := r.LoadByGVR(gvr); err == nil {
					pe.Namespaced = rd.Spec.Resource.Scope == v1alpha1.NamespaceScoped
					pe.Icons = rd.Spec.Icons
				}
			}
			section.Entries = append(section.Entries, pe)
		}
		sections[group] = section
	}

	// now, add all known types for apiGroup specific sections
	r.Visit(func(_ string, rd *v1alpha1.ResourceDescriptor) {
		if rd.Spec.Resource.IsOfficialType() {
			return // skip k8s.io api groups
		}

		section, found := sections[rd.Spec.Resource.Group]
		if !found {
			return
		}

		gvr := rd.Spec.Resource.GroupVersionResource()
		if _, found = existingGVRs[gvr]; !found {
			section.Entries = append(section.Entries, v1alpha1.PanelEntry{
				Entry: v1alpha1.Entry{
					Name: rd.Spec.Resource.Kind,
					Type: &v1alpha1.GroupVersionResource{
						Group:    rd.Spec.Resource.Group,
						Version:  rd.Spec.Resource.Version,
						Resource: rd.Spec.Resource.Name,
					},
					Icons: rd.Spec.Icons,
				},
				Namespaced: rd.Spec.Resource.Scope == v1alpha1.NamespaceScoped,
			})
			existingGVRs[gvr] = true
		}
	})

	// now, auto discover sections from CRDs
	apiext, err := crd_cs.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	crds, err := apiext.CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, crd := range crds.Items {
		group := crd.Spec.Group
		version := crd.Spec.Version
		for _, v := range crd.Spec.Versions {
			if v.Storage {
				version = v.Name
				break
			}
		}
		gvr := schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: crd.Spec.Names.Plural,
		}
		if _, found := existingGVRs[gvr]; found {
			continue
		}

		section, found := sections[group]
		if !found {
			if rc, found := KnownClasses[group]; found {
				w := math.MaxInt16
				if rc.Spec.Weight > 0 {
					w = rc.Spec.Weight
				}
				section = &v1alpha1.PanelSection{
					Name:              rc.Name,
					ResourceClassInfo: rc.Spec.ResourceClassInfo,
					Weight:            w,
				}
			} else {
				// unknown api group, so use CRD icon
				name := resourceclasses.ResourceClassName(group)
				section = &v1alpha1.PanelSection{
					Name: name,
					ResourceClassInfo: v1alpha1.ResourceClassInfo{
						APIGroup: group,
					},
					Weight: math.MaxInt16,
				}
			}
			sections[group] = section
		}

		section.Entries = append(section.Entries, v1alpha1.PanelEntry{
			Entry: v1alpha1.Entry{
				Name: crd.Spec.Names.Kind,
				Type: &v1alpha1.GroupVersionResource{
					Group:    group,
					Version:  version,
					Resource: crd.Spec.Names.Plural,
				},
			},
			Namespaced: crd.Spec.Scope == crdv1beta1.NamespaceScoped,
		})
		existingGVRs[gvr] = true
	}

	return toPanel(sections)
}

func toPanel(in map[string]*v1alpha1.PanelSection) (*v1alpha1.ResourcePanel, error) {
	sections := make([]*v1alpha1.PanelSection, 0, len(in))

	for key, section := range in {
		if !strings.HasSuffix(key, ".local") {
			sort.Slice(section.Entries, func(i, j int) bool {
				return section.Entries[i].Name < section.Entries[j].Name
			})
		}

		// Set icon for sections missing icon
		if len(section.Icons) == 0 {
			// TODO: Use a different icon for section
			section.Icons = []v1alpha1.ImageSpec{
				{
					Source: "https://cdn.appscode.com/k8s/icons/apiextensions.k8s.io/crd.svg",
					Type:   "image/svg+xml",
				},
			}
		}
		// set icons for entries missing icon
		for i := range section.Entries {
			if len(section.Entries[i].Icons) == 0 {
				section.Entries[i].Icons = []v1alpha1.ImageSpec{
					{
						Source: "https://cdn.appscode.com/k8s/icons/apiextensions.k8s.io/crd.svg",
						Type:   "image/svg+xml",
					},
				}
			}
		}

		sections = append(sections, section)
	}

	sort.Slice(sections, func(i, j int) bool {
		if sections[i].Weight == sections[j].Weight {
			return sections[i].Name < sections[j].Name
		}
		return sections[i].Weight < sections[j].Weight
	})

	return &v1alpha1.ResourcePanel{Sections: sections}, nil
}

type UnregisteredErr struct {
	t string
}

var _ error = UnregisteredErr{}

func (e UnregisteredErr) Error() string {
	return e.t + " isn't registered"
}
