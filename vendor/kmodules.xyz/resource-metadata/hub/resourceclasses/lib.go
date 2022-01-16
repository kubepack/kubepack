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

package resourceclasses

import (
	"embed"
	iofs "io/fs"
	"reflect"
	"strings"

	"kmodules.xyz/resource-metadata/apis/meta/v1alpha1"

	"github.com/gobuffalo/flect"
	"github.com/pkg/errors"
	"golang.org/x/net/publicsuffix"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

//go:embed **/*.yaml
var fs embed.FS

func FS() embed.FS {
	return fs
}

type UINamespace string

const (
	ClusterUI UINamespace = "cluster"
	KubeDBUI  UINamespace = "kubedb"
)

var (
	KnownClasses = map[UINamespace]map[string]*v1alpha1.ResourceClass{}
)

func init() {
	if e2 := iofs.WalkDir(FS(), ".", func(path string, e iofs.DirEntry, err error) error {
		if e.IsDir() || err != nil {
			return err
		}
		data, err := fs.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, path)
		}
		var rc v1alpha1.ResourceClass
		err = yaml.Unmarshal(data, &rc)
		if err != nil {
			return errors.Wrap(err, path)
		}

		if _, ok := KnownClasses[UINamespace(rc.Namespace)]; !ok {
			KnownClasses[UINamespace(rc.Namespace)] = map[string]*v1alpha1.ResourceClass{}
		}

		if rc.Spec.APIGroup != "" {
			KnownClasses[UINamespace(rc.Namespace)][rc.Spec.APIGroup] = &rc
		} else {
			KnownClasses[UINamespace(rc.Namespace)][strings.ToLower(rc.Name)+".local"] = &rc
		}
		return err
	}); e2 != nil {
		panic(errors.Wrapf(e2, "failed to load %s", reflect.TypeOf(v1alpha1.ResourceClass{})))
	}
}

func ResourceClassName(apiGroup string) string {
	name := apiGroup
	name = strings.TrimSuffix(name, ".k8s.io")
	name = strings.TrimSuffix(name, ".x-k8s.io")

	idx := strings.IndexRune(name, '.')
	if idx >= 0 {
		eTLD, icann := publicsuffix.PublicSuffix(name)
		if icann {
			name = strings.TrimSuffix(name, "."+eTLD)
		}
		parts := strings.Split(name, ".")
		for i := 0; i < len(parts)/2; i++ {
			j := len(parts) - i - 1
			parts[i], parts[j] = parts[j], parts[i]
		}
		name = strings.Join(parts, " ")
	}
	if name != "" {
		name = flect.Titleize(flect.Humanize(flect.Singularize(name)))
	} else {
		name = "Core"
	}
	return name
}

func LoadByGVR(namespace UINamespace, gvr schema.GroupVersionResource) (*v1alpha1.ResourceClass, error) {
	name := ResourceClassName(gvr.Group)
	return LoadByName(namespace, name)
}

func LoadByName(namespace UINamespace, name string) (*v1alpha1.ResourceClass, error) {
	name = strings.ToLower(name)
	if rcs, ok := KnownClasses[namespace]; ok {
		if obj, ok := rcs[name]; ok {
			return obj, nil
		}
	}
	return nil, apierrors.NewNotFound(v1alpha1.Resource(v1alpha1.ResourceKindResourceClass), name)
}
