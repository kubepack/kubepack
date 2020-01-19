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

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/util"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

var (
	file = "artifacts/kubedb-bundle/order.yaml"
)

func main() {
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.Parse()

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	var order v1alpha1.Order
	err = yaml.Unmarshal(data, &order)
	if err != nil {
		log.Fatal(err)
	}

	uid := uuid.New()

	var buf bytes.Buffer
	_, err = buf.WriteString("#!/usr/bin/env bash\n")
	if err != nil {
		log.Fatal(err)
	}
	_, err = buf.WriteString("set -xeou pipefail\n\n")
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		f1 := &util.NamespacePrinter{Namespace: pkg.Chart.Namespace, W: &buf}
		err = f1.Do()
		if err != nil {
			log.Fatal(err)
		}

		f2 := &util.Helm3CommandPrinter{
			ChartRef:    pkg.Chart.ChartRef,
			Version:     pkg.Chart.Version,
			ReleaseName: pkg.Chart.ReleaseName,
			Namespace:   pkg.Chart.Namespace,
			ValuesPatch: pkg.Chart.ValuesPatch,
			W:           &buf,
		}
		err = f2.Do()
		if err != nil {
			log.Fatal(err)
		}

		f3 := &util.WaitForPrinter{
			Name:      pkg.Chart.ReleaseName,
			Namespace: pkg.Chart.Namespace,
			WaitFors:  pkg.Chart.WaitFors,
			W:         &buf,
		}
		err = f3.Do()
		if err != nil {
			log.Fatal(err)
		}

		if pkg.Chart.Resources != nil && len(pkg.Chart.Resources.Owned) > 0 {
			f4 := &util.CRDReadinessPrinter{
				CRDs: pkg.Chart.Resources.Owned,
				W:    &buf,
			}
			err = f4.Do()
		}

		_, err = buf.WriteRune('\n')
		if err != nil {
			log.Fatal(err)
		}
	}

	err = util.Upload(uid.String(), "run.sh", buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())

	fmt.Printf("curl -fsSL %s/%s/run.sh  | bash", util.YAMLHost, uid)
}
