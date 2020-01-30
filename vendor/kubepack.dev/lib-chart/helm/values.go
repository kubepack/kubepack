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

package helm

import (
	"bytes"
	"io/ioutil"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	v1 "k8s.io/api/authorization/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubernetes/pkg/apis/core"
)

func GetDefaultValues(chartName, repoName, url string) (map[string]interface{}, error) {
	cfg := new(action.Configuration)
	client := action.NewInstall(cfg)

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
	err = getRepo(repoName, url)
	if err != nil {
		return nil, err
	}

	cp, err := client.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chartRequested.Values, nil
}

func GetManifests(chartName, repoName, url, name, namespace string, values []string, factory genericclioptions.RESTClientGetter) ([]string, error) {
	return getChart(chartName, repoName, url, name, namespace, values, factory, false)
}

func GetCreatePermission(manifest string, factory genericclioptions.RESTClientGetter) (bool, error) {
	cfg := new(action.Configuration)
	if err := cfg.Init(factory, core.NamespaceDefault, "", debug); err != nil {
		os.Exit(1)
	}
	resList, err := cfg.KubeClient.Build(bytes.NewBufferString(manifest), true)
	if err != nil {
		return false, err
	}
	res := resList[0]
	sar := v1.SelfSubjectAccessReview{
		Spec: v1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				Namespace: res.Namespace,
				Verb:      "create",
				Group:     res.Mapping.Resource.Group,
				Version:   res.Mapping.Resource.Version,
				Resource:  res.Mapping.Resource.Resource,
			},
		},
	}
	restConfig, err := factory.ToRESTConfig()
	if err != nil {
		return false, err
	}
	return isSARAllowed(sar, restConfig)
}
