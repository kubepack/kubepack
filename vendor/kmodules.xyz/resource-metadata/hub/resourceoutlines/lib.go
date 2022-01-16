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

package resourceoutlines

import (
	"embed"
	"fmt"
	iofs "io/fs"
	"reflect"
	"sort"

	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

//go:embed **/**/*.yaml **/**/**/*.yaml
var fs embed.FS

func FS() embed.FS {
	return fs
}

var (
	rlMap   = map[string]*v1alpha1.ResourceOutline{}
	rlPerGK = map[schema.GroupVersionKind]*v1alpha1.ResourceOutline{}
	rlPerGR = map[schema.GroupVersionResource]*v1alpha1.ResourceOutline{}
)

func init() {
	if err := iofs.WalkDir(fs, ".", func(path string, d iofs.DirEntry, err error) error {
		if d.IsDir() || err != nil {
			return err
		}
		data, err := fs.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, path)
		}
		var obj v1alpha1.ResourceOutline
		err = yaml.Unmarshal(data, &obj)
		if err != nil {
			return errors.Wrap(err, path)
		}
		rlMap[obj.Name] = &obj

		if obj.Spec.DefaultLayout {
			gvr := obj.Spec.Resource.GroupVersionResource()
			expectedName := DefaultLayoutName(gvr)
			if obj.Name != expectedName {
				return fmt.Errorf("expected default %s name to be %s, found %s", reflect.TypeOf(v1alpha1.ResourceOutline{}), expectedName, obj.Name)
			}

			gvk := obj.Spec.Resource.GroupVersionKind()
			if rv, ok := rlPerGK[gvk]; !ok {
				rlPerGK[gvk] = &obj
			} else {
				return fmt.Errorf("multiple %s found for %+v: %s and %s", reflect.TypeOf(v1alpha1.ResourceOutline{}), gvk, rv.Name, obj.Name)
			}
			if rv, ok := rlPerGR[gvr]; !ok {
				rlPerGR[gvr] = &obj
			} else {
				return fmt.Errorf("multiple %s found for %+v: %s and %s", reflect.TypeOf(v1alpha1.ResourceOutline{}), gvk, rv.Name, obj.Name)
			}
		}
		return nil
	}); err != nil {
		panic(errors.Wrapf(err, "failed to load %s", reflect.TypeOf(v1alpha1.ResourceOutline{})))
	}
}

func DefaultLayoutName(gvr schema.GroupVersionResource) string {
	if gvr.Group == "" && gvr.Version == "v1" {
		return fmt.Sprintf("core-v1-%s", gvr.Resource)
	}
	return fmt.Sprintf("%s-%s-%s", gvr.Group, gvr.Version, gvr.Resource)
}

func LoadByName(name string) (*v1alpha1.ResourceOutline, error) {
	if obj, ok := rlMap[name]; ok {
		return obj, nil
	}
	return nil, apierrors.NewNotFound(v1alpha1.Resource(v1alpha1.ResourceKindResourceOutline), name)
}

func DefaultOutlineForGVK(gvk schema.GroupVersionKind) (*v1alpha1.ResourceOutline, bool) {
	rv, found := rlPerGK[gvk]
	return rv, found
}

func DefaultOutlineForGVR(gvr schema.GroupVersionResource) (*v1alpha1.ResourceOutline, bool) {
	rv, found := rlPerGR[gvr]
	return rv, found
}

func List() []v1alpha1.ResourceOutline {
	out := make([]v1alpha1.ResourceOutline, 0, len(rlMap))
	for _, rl := range rlMap {
		out = append(out, *rl)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}
