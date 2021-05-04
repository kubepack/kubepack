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

package controller

import (
	"fmt"

	api "kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	cs "kubepack.dev/kubepack/client/clientset/versioned"

	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/tools/queue"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	app_cs "sigs.k8s.io/application/client/clientset/versioned"
	appinformers "sigs.k8s.io/application/client/informers/externalversions"
	app_listers "sigs.k8s.io/application/client/listers/app/v1beta1"
)

type KubepackController struct {
	config
	clientConfig *rest.Config

	kubeClient       kubernetes.Interface
	extClient        cs.Interface
	appClient        app_cs.Interface
	appCatalogClient appcat_cs.AppcatalogV1alpha1Interface
	crdClient        crd_cs.Interface
	recorder         record.EventRecorder
	// Prometheus client
	promClient pcm.MonitoringV1Interface

	kubeInformerFactory informers.SharedInformerFactory
	extInformerFactory  appinformers.SharedInformerFactory

	// for Application
	appQueue    *queue.Worker
	appInformer cache.SharedIndexInformer
	appLister   app_listers.ApplicationLister
}

func (c *KubepackController) ensureCustomResourceDefinitions() error {
	crds := []*apiextensions.CustomResourceDefinition{
		api.Bundle{}.CustomResourceDefinition(),
		api.Order{}.CustomResourceDefinition(),
		api.ApplicationCustomResourceDefinition(),
		appcat.AppBinding{}.CustomResourceDefinition(),
	}
	return apiextensions.RegisterCRDs(c.crdClient, crds)
}

func (c *KubepackController) Run(stopCh <-chan struct{}) {
	go c.RunInformers(stopCh)

	if c.EnableMutatingWebhook {
		cancel, _ := reg_util.SyncMutatingWebhookCABundle(c.clientConfig, mutatingWebhook)
		defer cancel()
	}
	if c.EnableValidatingWebhook {
		cancel, _ := reg_util.SyncValidatingWebhookCABundle(c.clientConfig, validatingWebhook)
		defer cancel()
	}

	<-stopCh
}

func (c *KubepackController) RunInformers(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	klog.Info("Starting Vault controller")

	c.extInformerFactory.Start(stopCh)
	for _, v := range c.extInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}
	}

	//For Application
	go c.appQueue.Run(stopCh)

	<-stopCh
	klog.Info("Stopping Vault operator")
}
