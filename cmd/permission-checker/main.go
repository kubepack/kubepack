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
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	clientcmdutil "kmodules.xyz/client-go/tools/clientcmd"
	"kmodules.xyz/resource-metadata/hub"
	"sigs.k8s.io/yaml"
)

var (
	masterURL      = ""
	kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	file           = "artifacts/kubedb-community/order.yaml"
)

func main() {
	flag.StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	flag.StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&file, "file", file, "Path to Order file")
	flag.Parse()

	if masterURL == "" && kubeconfigPath == "" {
		log.Fatalln("Possibly in cluster. Can't create RESTClientGetter")
	}

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}})
	kubeconfig, err := cc.RawConfig()
	if err != nil {
		log.Fatal(err)
	}
	getter := clientcmdutil.NewClientGetter(&kubeconfig)

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	var order v1alpha1.Order
	err = yaml.Unmarshal(data, &order)
	if err != nil {
		log.Fatal(err)
	}
	order.UID = types.UID(uuid.New().String())

	allowed, err := lib.CheckPermissions(getter, lib.DefaultRegistry, order, hub.Helm3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ALLOWED", "=", allowed)
}
