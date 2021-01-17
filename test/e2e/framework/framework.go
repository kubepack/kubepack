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

package framework

import (
	cs "kubepack.dev/kubepack/client/clientset/versioned"

	"gomodules.xyz/x/crypto/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	app_cs "sigs.k8s.io/application/client/clientset/versioned"
)

type Framework struct {
	restConfig *rest.Config
	kubeClient kubernetes.Interface
	client     cs.Interface
	appClient  app_cs.Interface
	namespace  string
	name       string
}

func New(
	restConfig *rest.Config,
	kubeClient kubernetes.Interface,
	client cs.Interface,
	appClient app_cs.Interface,
) *Framework {
	return &Framework{
		restConfig: restConfig,
		kubeClient: kubeClient,
		client:     client,
		appClient:  appClient,
		name:       "kfc",
		namespace:  rand.WithUniqSuffix("kubepack"),
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("kp-e2e"),
	}
}

func (fi *Invocation) KubeClient() kubernetes.Interface {
	return fi.kubeClient
}

func (fi *Invocation) RestConfig() *rest.Config {
	return fi.restConfig
}

func (fi *Invocation) KubeformClient() cs.Interface {
	return fi.client
}

type Invocation struct {
	*Framework
	app string
}
