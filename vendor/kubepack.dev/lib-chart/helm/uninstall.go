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
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func UninstallChart(name string, namespace string, factory genericclioptions.RESTClientGetter) error {
	cfg := new(action.Configuration)

	if err := cfg.Init(factory, namespace, "", debug); err != nil {
		return err
	}

	cfg.Capabilities = chartutil.DefaultCapabilities

	client := action.NewUninstall(cfg)
	_, err := client.Run(name)

	return err
}
