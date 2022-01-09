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
	"reflect"
	"sort"

	"kmodules.xyz/apiversion"
	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

//go:embed **/**/*.yaml
var fs embed.FS

func FS() embed.FS {
	return fs
}

var (
	KnownDescriptors = make(map[string]*v1alpha1.ResourceDescriptor)
	LatestGVRs       = make(map[schema.GroupKind]schema.GroupVersionResource)
)

func init() {
	e2 := iofs.WalkDir(FS(), ".", func(path string, e iofs.DirEntry, err error) error {
		if e.IsDir() || err != nil {
			return err
		}
		data, err := fs.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, path)
		}
		var rd v1alpha1.ResourceDescriptor
		err = yaml.Unmarshal(data, &rd)
		if err != nil {
			return errors.Wrap(err, path)
		}
		KnownDescriptors[rd.Name] = &rd

		gvr := rd.Spec.Resource.GroupVersionResource()
		gk := rd.Spec.Resource.GroupKind()
		if existing, ok := LatestGVRs[gk]; !ok {
			LatestGVRs[gk] = gvr
		} else if diff, _ := apiversion.Compare(existing.Version, gvr.Version); diff < 0 {
			LatestGVRs[gk] = gvr
		}
		return err
	})
	if e2 != nil {
		panic(errors.Wrapf(e2, "failed to load %s", reflect.TypeOf(v1alpha1.ResourceDescriptor{})))
	}
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
	if obj, ok := KnownDescriptors[name]; ok {
		return obj, nil
	}
	return nil, apierrors.NewNotFound(v1alpha1.Resource(v1alpha1.ResourceKindResourceDescriptor), name)
}

func List() []v1alpha1.ResourceDescriptor {
	out := make([]v1alpha1.ResourceDescriptor, 0, len(KnownDescriptors))
	for _, rl := range KnownDescriptors {
		out = append(out, *rl)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}
