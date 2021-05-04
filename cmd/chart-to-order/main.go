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

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	url     = "https://bundles.byte.builders/ui/"
	name    = "mongodb-editor-options"
	version = "v0.1.0"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	order := v1alpha1.Order{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindOrder,
		}, ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			UID:               types.UID(uuid.New().String()),
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: v1alpha1.OrderSpec{
			Packages: []v1alpha1.PackageSelection{
				{
					Chart: &v1alpha1.ChartSelection{
						ChartRef: v1alpha1.ChartRef{
							URL:  url,
							Name: name,
						},
						Version:     version,
						ReleaseName: name,
						Namespace:   metav1.NamespaceDefault,
						Bundle:      nil,
						ValuesFile:  "values.yaml",
						ValuesPatch: nil,
						Resources:   nil,
						WaitFors:    nil,
					},
				},
			},
			KubeVersion: "",
		},
	}

	err := os.MkdirAll("artifacts/"+name, 0755)
	if err != nil {
		klog.Fatal(err)
	}

	{
		data, err := yaml.Marshal(order)
		if err != nil {
			klog.Fatal(err)
		}
		err = ioutil.WriteFile("artifacts/"+name+"/order.yaml", data, 0644)
		if err != nil {
			klog.Fatal(err)
		}
	}

	{
		data, err := json.MarshalIndent(order, "", "  ")
		if err != nil {
			klog.Fatal(err)
		}
		err = ioutil.WriteFile("artifacts/"+name+"/order.json", data, 0644)
		if err != nil {
			klog.Fatal(err)
		}
	}
}
