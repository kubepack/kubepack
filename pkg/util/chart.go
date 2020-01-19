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
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/gabriel-vasile/mimetype"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/helmpath/xdg"
)

const (
	TmpDir    = "/tmp"
	DirPrefix = "helm"
)

func GetPackageDescriptor(pkgChart *chart.Chart) v1alpha1.PackageDescriptor {
	var out v1alpha1.PackageDescriptor

	out.Description = pkgChart.Metadata.Description
	if pkgChart.Metadata.Icon != "" {
		var imgType string
		if resp, err := http.Get(pkgChart.Metadata.Icon); err == nil {
			if mime, err := mimetype.DetectReader(resp.Body); err == nil {
				imgType = mime.String()
			}
			resp.Body.Close()
		}
		out.Icons = []v1alpha1.ImageSpec{
			{
				Source: pkgChart.Metadata.Icon,
				// TotalSize: "",
				Type: imgType,
			},
		}
	}
	for _, maintainer := range pkgChart.Metadata.Maintainers {
		out.Maintainers = append(out.Maintainers, v1alpha1.ContactData{
			Name:  maintainer.Name,
			URL:   maintainer.URL,
			Email: maintainer.Email,
		})
	}
	out.Keywords = pkgChart.Metadata.Keywords

	if pkgChart.Metadata.Home != "" {
		out.Links = []v1alpha1.Link{
			{
				Description: "Homepage",
				URL:         pkgChart.Metadata.Home,
			},
		}
	}

	return out
}

func GetChart(chartName, version, repoName, url string) (*chart.Chart, error) {
	cfg := new(action.Configuration)
	client := action.NewInstall(cfg)
	client.Version = version
	client.RepoURL = url

	chartDir, err := ioutil.TempDir(TmpDir, DirPrefix)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(chartDir)

	err = setEnv(chartDir)
	if err != nil {
		return nil, err
	}
	defer unsetEnv()

	settings := cli.New()
	cp, err := client.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chartRequested, nil
}

func setEnv(chartDir string) error {
	err := os.Setenv(xdg.CacheHomeEnvVar, filepath.Join(chartDir, "cache"))
	if err != nil {
		return err
	}

	err = os.Setenv(xdg.ConfigHomeEnvVar, filepath.Join(chartDir, "Config"))
	if err != nil {
		return err
	}

	return os.Setenv(xdg.DataHomeEnvVar, filepath.Join(chartDir, "data"))
}

func unsetEnv() error {
	err := os.Unsetenv(xdg.CacheHomeEnvVar)
	if err != nil {
		return err
	}

	err = os.Unsetenv(xdg.ConfigHomeEnvVar)
	if err != nil {
		return err
	}

	return os.Unsetenv(xdg.DataHomeEnvVar)
}
