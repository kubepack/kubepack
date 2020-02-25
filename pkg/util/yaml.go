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

package util

import (
	"bytes"
	"fmt"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"k8s.io/apimachinery/pkg/util/sets"
)

func GenerateYAMLScript(order v1alpha1.Order) (string, error) {
	var buf bytes.Buffer
	_, err := buf.WriteString("#!/usr/bin/env bash\n")
	if err != nil {
		return "", err
	}
	_, err = buf.WriteString("set -xeou pipefail\n\n")
	if err != nil {
		return "", err
	}

	namespaces := sets.NewString("default", "kube-system")

	f1 := &ApplicationCRDRegPrinter{
		W: &buf,
	}
	err = f1.Do()
	if err != nil {
		return "", err
	}
	_, err = buf.WriteRune('\n')
	if err != nil {
		return "", err
	}

	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		if !namespaces.Has(pkg.Chart.Namespace) {
			f2 := &NamespacePrinter{
				Namespace: pkg.Chart.Namespace,
				W:         &buf,
			}
			err = f2.Do()
			if err != nil {
				return "", err
			}
			namespaces.Insert(pkg.Chart.Namespace)
		}

		f3 := &YAMLPrinter{
			ChartRef:    pkg.Chart.ChartRef,
			Version:     pkg.Chart.Version,
			ReleaseName: pkg.Chart.ReleaseName,
			Namespace:   pkg.Chart.Namespace,
			KubeVersion: "v1.17.0",
			ValuesPatch: pkg.Chart.ValuesPatch,
			BucketURL:   YAMLBucket,
			UID:         string(order.UID),
			PublicURL:   YAMLHost,
			W:           &buf,
		}
		err = f3.Do()
		if err != nil {
			return "", err
		}

		f4 := &WaitForPrinter{
			Name:      pkg.Chart.ReleaseName,
			Namespace: pkg.Chart.Namespace,
			WaitFors:  pkg.Chart.WaitFors,
			W:         &buf,
		}
		err = f4.Do()
		if err != nil {
			return "", err
		}

		if pkg.Chart.Resources != nil && len(pkg.Chart.Resources.Owned) > 0 {
			f5 := &CRDReadinessPrinter{
				CRDs: pkg.Chart.Resources.Owned,
				W:    &buf,
			}
			err = f5.Do()
			if err != nil {
				return "", err
			}
		}

		f6 := &ApplicationGenerator{
			Chart:       *pkg.Chart,
			KubeVersion: "v1.17.0",
		}
		err = f6.Do()
		if err != nil {
			return "", err
		}

		f7 := &ApplicationUploader{
			App:       f6.Result(),
			UID:       string(order.UID),
			BucketURL: YAMLBucket,
			PublicURL: YAMLHost,
			W:         &buf,
		}
		err = f7.Do()
		if err != nil {
			return "", err
		}

		_, err = buf.WriteRune('\n')
		if err != nil {
			return "", err
		}
	}

	err = Upload(string(order.UID), "run.sh", buf.Bytes())
	if err != nil {
		return "", err
	}

	fmt.Println(buf.String())

	return fmt.Sprintf("curl -fsSL %s/%s/run.sh  | bash", YAMLHost, order.UID), nil
}
