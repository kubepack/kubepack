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

package lib

import (
	"bytes"
	"context"
	"fmt"
	"path"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/lib-helm/repo"

	"k8s.io/apimachinery/pkg/util/sets"
)

func GenerateHelm2Script(bs *BlobStore, reg *repo.Registry, order v1alpha1.Order) ([]ScriptRef, error) {
	var buf bytes.Buffer
	_, err := buf.WriteString("#!/usr/bin/env bash\n")
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteString("set -xeou pipefail\n\n")
	if err != nil {
		return nil, err
	}

	namespaces := sets.NewString("default", "kube-system")

	f1 := &ApplicationCRDRegPrinter{
		W: &buf,
	}
	err = f1.Do()
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteRune('\n')
	if err != nil {
		return nil, err
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
				return nil, err
			}
			namespaces.Insert(pkg.Chart.Namespace)
		}

		f3 := &Helm2CommandPrinter{
			Registry:    reg,
			ChartRef:    pkg.Chart.ChartRef,
			Version:     pkg.Chart.Version,
			ReleaseName: pkg.Chart.ReleaseName,
			Namespace:   pkg.Chart.Namespace,
			ValuesFile:  pkg.Chart.ValuesFile,
			ValuesPatch: pkg.Chart.ValuesPatch,
			W:           &buf,
		}
		err = f3.Do()
		if err != nil {
			return nil, err
		}

		f4 := &WaitForPrinter{
			Name:      pkg.Chart.ReleaseName,
			Namespace: pkg.Chart.Namespace,
			WaitFors:  pkg.Chart.WaitFors,
			W:         &buf,
		}
		err = f4.Do()
		if err != nil {
			return nil, err
		}

		if pkg.Chart.Resources != nil && len(pkg.Chart.Resources.Owned) > 0 {
			f5 := &CRDReadinessPrinter{
				CRDs: pkg.Chart.Resources.Owned,
				W:    &buf,
			}
			err = f5.Do()
			if err != nil {
				return nil, err
			}
		}

		f6 := &ApplicationGenerator{
			Registry:    reg,
			Chart:       *pkg.Chart,
			KubeVersion: "v1.17.0",
		}
		err = f6.Do()
		if err != nil {
			return nil, err
		}

		f7 := &ApplicationUploader{
			App:       f6.Result(),
			UID:       string(order.UID),
			BucketURL: bs.Bucket,
			PublicURL: bs.Host,
			W:         &buf,
		}
		err = f7.Do()
		if err != nil {
			return nil, err
		}

		_, err = buf.WriteRune('\n')
		if err != nil {
			return nil, err
		}
	}

	err = bs.WriteFile(context.TODO(), path.Join(string(order.UID), "helm2.sh"), buf.Bytes())
	if err != nil {
		return nil, err
	}

	scriptURL := fmt.Sprintf("%s/%s/helm2.sh", bs.Host, order.UID)
	return []ScriptRef{
		{
			OS:      Linux,
			URL:     scriptURL,
			Command: fmt.Sprintf("curl -fsSL %s | bash", scriptURL),
		},
		{
			OS:      MacOS,
			URL:     scriptURL,
			Command: fmt.Sprintf("curl -fsSL %s | bash", scriptURL),
		},
	}, nil
}
