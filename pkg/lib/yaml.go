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
	"strings"

	"kubepack.dev/lib-helm/pkg/repo"

	"x-helm.dev/apimachinery/apis"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

func GenerateYAMLScript(bs *BlobStore, reg repo.IRegistry, order releasesapi.Order, opts ...ScriptOption) ([]ScriptRef, error) {
	var buf bytes.Buffer
	var err error

	var scriptOptions ScriptOptions
	for _, opt := range opts {
		opt.Apply(&scriptOptions)
	}

	if !scriptOptions.OsIndependentScript {
		_, err = buf.WriteString("#!/usr/bin/env sh\n")
		if err != nil {
			return nil, err
		}
	}

	if !scriptOptions.DisableApplicationCRD {
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
	}

	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		f3 := &YAMLPrinter{
			Registry:    reg,
			ChartRef:    pkg.Chart.ChartRef,
			Version:     pkg.Chart.Version,
			ReleaseName: pkg.Chart.ReleaseName,
			Namespace:   pkg.Chart.Namespace,
			KubeVersion: apis.DefaultKubernetesVersion,
			ValuesFile:  pkg.Chart.ValuesFile,
			ValuesPatch: pkg.Chart.ValuesPatch,
			BucketURL:   bs.Bucket,
			UID:         string(order.UID),
			PublicURL:   bs.Host,
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

		if !scriptOptions.DisableApplicationCRD {
			f6 := &ApplicationGenerator{
				Registry:    reg,
				Chart:       *pkg.Chart,
				KubeVersion: apis.DefaultKubernetesVersion,
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
		}

		_, err = buf.WriteRune('\n')
		if err != nil {
			return nil, err
		}
	}

	if scriptOptions.OsIndependentScript {
		return []ScriptRef{
			{
				OS:      Neutral,
				URL:     "",
				Command: "",
				Script:  strings.TrimSpace(buf.String()),
			},
		}, nil
	}

	err = bs.WriteFile(context.TODO(), path.Join(string(order.UID), "script.sh"), buf.Bytes())
	if err != nil {
		return nil, err
	}

	scriptURL := fmt.Sprintf("%s/%s/script.sh", bs.Host, order.UID)
	return []ScriptRef{
		{
			OS:      Linux,
			URL:     scriptURL,
			Command: fmt.Sprintf("curl -fsSL %s | bash", scriptURL),
			Script:  buf.String(),
		},
		{
			OS:      MacOS,
			URL:     scriptURL,
			Command: fmt.Sprintf("curl -fsSL %s | bash", scriptURL),
			Script:  buf.String(),
		},
	}, nil
}
