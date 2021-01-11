/*
Copyright The Kubepack Authors.

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

package lib

import (
	"bytes"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type ResourceFn func(obj *unstructured.Unstructured) error

func ProcessResources(data []byte, fn ResourceFn) error {
	reader := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 2048)
	for {
		var obj unstructured.Unstructured
		err := reader.Decode(&obj)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if obj.IsList() {
			if err := obj.EachListItem(func(item runtime.Object) error {
				return fn(item.(*unstructured.Unstructured))
			}); err != nil {
				return err
			}
		} else {
			if err := fn(&obj); err != nil {
				return err
			}
		}
	}
	return nil
}
