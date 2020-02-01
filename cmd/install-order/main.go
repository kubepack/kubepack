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
	"io/ioutil"
	"log"
	"path/filepath"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/client/clientset/versioned"
	"kubepack.dev/kubepack/pkg/util"

	flag "github.com/spf13/pflag"
	"gomodules.xyz/version"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	clientcmdutil "kmodules.xyz/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

var (
	masterURL      = ""
	kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	file           = "artifacts/kubedb-bundle/order.yaml"
)

func main() {
	flag.StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	flag.StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
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

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}})
	kubeconfig, err := cc.RawConfig()
	if err != nil {
		log.Fatal(err)
	}
	getter := clientcmdutil.NewClientGetter(&kubeconfig)

	config, err := cc.ClientConfig()
	if err != nil {
		log.Fatalln(err)
	}

	kc, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	info, err := kc.ServerVersion()
	if err != nil {
		log.Fatal(err)
	}
	kv, err := version.NewVersion(info.GitVersion)
	if err != nil {
		log.Fatal(err)
	}
	kubeVersion := kv.ToMutator().ResetPrerelease().ResetMetadata().Done().String()

	namespaces := sets.NewString("default", "kube-system")

	f1 := &util.ApplicationCRDRegistrar{
		Config: config,
	}
	err = f1.Do()
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		if !namespaces.Has(pkg.Chart.Namespace) {
			f2 := &util.NamespaceCreator{
				Namespace: pkg.Chart.Namespace,
				Client:    kc,
			}
			err = f2.Do()
			if err != nil {
				log.Fatal(err)
			}
			namespaces.Insert(pkg.Chart.Namespace)
		}

		f3 := &util.ChartInstaller{
			ChartRef:     pkg.Chart.ChartRef,
			Version:      pkg.Chart.Version,
			ReleaseName:  pkg.Chart.ReleaseName,
			Namespace:    pkg.Chart.Namespace,
			ValuesPatch:  pkg.Chart.ValuesPatch,
			ClientGetter: getter,
		}
		err = f3.Do()
		if err != nil {
			log.Fatal(err)
		}

		f4 := &util.WaitForChecker{
			Namespace:    pkg.Chart.Namespace,
			WaitFors:     pkg.Chart.WaitFors,
			ClientGetter: getter,
		}
		err = f4.Do()
		if err != nil {
			log.Fatal(err)
		}

		if pkg.Chart.Resources != nil && len(pkg.Chart.Resources.Owned) > 0 {
			f5 := &util.CRDReadinessChecker{
				CRDs:   pkg.Chart.Resources.Owned,
				Client: kc.RESTClient(),
			}
			err = f5.Do()
			if err != nil {
				log.Fatal(err)
			}
		}

		f6 := &util.ApplicationGenerator{
			Chart:       *pkg.Chart,
			KubeVersion: kubeVersion,
		}
		err = f6.Do()
		if err != nil {
			log.Fatal(err)
		}

		f7 := &util.ApplicationCreator{
			App:    f6.Result(),
			Client: versioned.NewForConfigOrDie(config),
		}
		err = f7.Do()
		if err != nil {
			log.Fatal(err)
		}
	}
}