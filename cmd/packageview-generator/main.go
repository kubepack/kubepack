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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/util"

	flag "github.com/spf13/pflag"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	yamllib "sigs.k8s.io/yaml"
)

var (
	//url     = "https://charts.appscode.com/stable/"
	//name    = "kubedb"
	//version = "v0.13.0-rc.0"

	url     = "https://kubepack-testcharts.storage.googleapis.com"
	name    = "stash"
	version = "v0.9.0-rc.2"
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	pkgChart, err := util.GetChart(url, name, version)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(pkgChart.Metadata.Description)

	b := v1alpha1.PackageView{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "PackageView",
		},
		PackageMeta: v1alpha1.PackageMeta{
			Name:              pkgChart.Name(),
			URL:               url,
			Version:           pkgChart.Metadata.Version,
			PackageDescriptor: util.GetPackageDescriptor(pkgChart.Chart),
		},
		Values: &runtime.RawExtension{
			Object: &unstructured.Unstructured{Object: pkgChart.Values},
		},
	}

	for _, f := range pkgChart.Files {
		if f.Name == "values.openapiv3_schema.json" || f.Name == "values.openapiv3_schema.yaml" || f.Name == "values.openapiv3_schema.yml" {
			var schema crdv1beta1.JSONSchemaProps
			reader := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(f.Data), 2048)
			err := reader.Decode(&schema)
			if err != nil {
				log.Fatalln(err)
			}
			b.OpenAPIV3Schema = &schema
		}
	}
	//if b.Schema == nil && len(pkgChart.Schema) > 0 {
	//	// TODO convert json schema to openapi schema v3
	//}

	data, err := yamllib.Marshal(b)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("artifacts/"+name, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("artifacts/"+name+"/packageview.yaml", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
