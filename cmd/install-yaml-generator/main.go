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
	"fmt"
	"os"

	"kubepack.dev/kubepack/cmd/internal"
	"kubepack.dev/kubepack/pkg/lib"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

var file = "artifacts/kubedb-community/order.yaml"

func main() {
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.Parse()

	bs, err := lib.NewTestBlobStore()
	if err != nil {
		klog.Fatal(err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatal(err)
	}
	var order releasesapi.Order
	err = yaml.Unmarshal(data, &order)
	if err != nil {
		klog.Fatal(err)
	}
	order.UID = types.UID(uuid.New().String())

	scripts, err := lib.GenerateYAMLScript(bs, internal.DefaultRegistry, order)
	if err != nil {
		klog.Fatal(err)
	}
	data, err = json.MarshalIndent(scripts, "", "  ")
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println(string(data))
}
