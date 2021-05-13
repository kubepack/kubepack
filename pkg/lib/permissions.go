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
	"fmt"
	"os"
	"text/tabwriter"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/lib-helm/repo"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/restmapper"
	disco_util "kmodules.xyz/client-go/discovery"
	"kmodules.xyz/resource-metadata/hub"
)

func CheckPermissions(getter genericclioptions.RESTClientGetter, reg *repo.Registry, order v1alpha1.Order, helm hub.HelmVersion) (bool, error) {
	config, err := getter.ToRESTConfig()
	if err != nil {
		return false, err
	}
	discoclient, err := getter.ToDiscoveryClient()
	if err != nil {
		return false, err
	}
	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		// TODO: What does permission check mean for non-existent resources?
		checker := &PermissionChecker{
			Registry:    reg,
			ChartRef:    pkg.Chart.ChartRef,
			Version:     pkg.Chart.Version,
			ReleaseName: pkg.Chart.ReleaseName,
			Namespace:   pkg.Chart.Namespace,
			Verb:        "create",

			Config:       config,
			ClientGetter: getter,
			Mapper:       disco_util.NewResourceMapper(restmapper.NewDeferredDiscoveryRESTMapper(discoclient)),
		}
		err = checker.Do()
		if err != nil {
			return false, err
		}
		attrs, allowed := checker.Result()
		if !allowed {
			fmt.Println("Install not permitted")
			return false, nil
		}

		w := new(tabwriter.Writer)
		// Format in tab-separated columns with a tab stop of 8.
		w.Init(os.Stdout, 0, 20, 0, '\t', 0)
		fmt.Fprintln(w, "Group\tVersion\tResource\tNamespace\tName\tAllowed\t")
		for k, v := range attrs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%v\n", k.Group, k.Version, k.Resource, k.Namespace, k.Name, v.Allowed)
		}
		fmt.Fprintln(w)
		w.Flush()
	}
	return true, nil
}
