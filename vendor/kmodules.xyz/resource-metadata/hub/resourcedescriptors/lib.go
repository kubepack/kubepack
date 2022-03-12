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

package resourcedescriptors

import (
	"embed"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"sync"

	"kmodules.xyz/apiversion"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"

	"github.com/pkg/errors"
	ioutilx "gomodules.xyz/x/ioutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

var (
	//go:embed **/**/*.yaml trigger
	fs embed.FS

	m                sync.Mutex
	knownDescriptors map[string]*v1alpha1.ResourceDescriptor
	latestGVRs       map[schema.GroupKind]schema.GroupVersionResource

	loader = ioutilx.NewReloader(
		filepath.Join("/tmp", "hub", "resourcedescriptors"),
		fs,
		func(fsys iofs.FS) {
			knownDescriptors = map[string]*v1alpha1.ResourceDescriptor{}
			latestGVRs = map[schema.GroupKind]schema.GroupVersionResource{}

			e2 := iofs.WalkDir(fsys, ".", func(path string, d iofs.DirEntry, err error) error {
				if d.IsDir() || err != nil {
					return errors.Wrap(err, path)
				}
				ext := filepath.Ext(d.Name())
				if ext != ".yaml" && ext != ".yml" && ext != ".json" {
					return nil
				}

				data, err := iofs.ReadFile(fsys, path)
				if err != nil {
					return errors.Wrap(err, path)
				}
				var rd v1alpha1.ResourceDescriptor
				err = yaml.Unmarshal(data, &rd)
				if err != nil {
					return errors.Wrap(err, path)
				}
				knownDescriptors[rd.Name] = &rd

				gvr := rd.Spec.Resource.GroupVersionResource()
				gk := rd.Spec.Resource.GroupKind()
				if existing, ok := latestGVRs[gk]; !ok {
					latestGVRs[gk] = gvr
				} else if diff, _ := apiversion.Compare(existing.Version, gvr.Version); diff < 0 {
					latestGVRs[gk] = gvr
				}
				return err
			})
			if e2 != nil {
				panic(errors.Wrapf(e2, "failed to load %s", reflect.TypeOf(v1alpha1.ResourceDescriptor{})))
			}
		},
	)
)

func init() {
	loader.ReloadIfTriggered()
}

func EmbeddedFS() iofs.FS {
	return fs
}

func KnownDescriptors() map[string]*v1alpha1.ResourceDescriptor {
	m.Lock()
	defer m.Unlock()
	loader.ReloadIfTriggered()

	return knownDescriptors
}

func LatestGVRs() map[schema.GroupKind]schema.GroupVersionResource {
	m.Lock()
	defer m.Unlock()
	loader.ReloadIfTriggered()

	return latestGVRs
}

func LoadByGVR(gvr schema.GroupVersionResource) (*v1alpha1.ResourceDescriptor, error) {
	return LoadByName(GetName(gvr))
}

func GetName(gvr schema.GroupVersionResource) string {
	if gvr.Group == "" && gvr.Version == "v1" {
		return fmt.Sprintf("core-v1-%s", gvr.Resource)
	}
	return fmt.Sprintf("%s-%s-%s", gvr.Group, gvr.Version, gvr.Resource)
}

func LoadByName(name string) (*v1alpha1.ResourceDescriptor, error) {
	m.Lock()
	defer m.Unlock()
	loader.ReloadIfTriggered()

	if obj, ok := knownDescriptors[name]; ok {
		return obj, nil
	}
	return nil, apierrors.NewNotFound(v1alpha1.Resource(v1alpha1.ResourceKindResourceDescriptor), name)
}

func List() []v1alpha1.ResourceDescriptor {
	m.Lock()
	defer m.Unlock()
	loader.ReloadIfTriggered()

	out := make([]v1alpha1.ResourceDescriptor, 0, len(knownDescriptors))
	for _, rl := range knownDescriptors {
		out = append(out, *rl)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func Names() []string {
	m.Lock()
	defer m.Unlock()
	loader.ReloadIfTriggered()

	out := make([]string, 0, len(knownDescriptors))
	for name := range knownDescriptors {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
