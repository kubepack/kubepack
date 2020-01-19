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

	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalln(err)
	}
	kc, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	nodes, err := kc.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(nodes.Items)

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	var bv v1alpha1.Order
	err = yaml.Unmarshal(data, &bv)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range bv.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}
		fmt.Println(pkg.Chart.ChartRef, pkg.Chart.Version)
	}
}
