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

package controller

import (
	api "kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/client/clientset/versioned/typed/kubepack/v1alpha1/util"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	AppFinalizer = "kubepack.dev"
)

func (c *KubepackController) initAppWatcher() {
	c.appInformer = c.extInformerFactory.Kubepack().V1alpha1().Applications().Informer()
	c.appQueue = queue.New(api.ResourceKindApplication, c.MaxNumRequeues, c.NumThreads, c.runAppInjector)
	c.appInformer.AddEventHandler(queue.NewReconcilableHandler(c.appQueue.GetQueue()))
	c.appLister = c.extInformerFactory.Kubepack().V1alpha1().Applications().Lister()
}

// runAppInjector gets the vault policy object indexed by the key from cache
// and initializes, reconciles or garbage collects the vault policy as needed.
func (c *KubepackController) runAppInjector(key string) error {
	obj, exists, err := c.appInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		glog.Warningf("Application %s does not exist anymore\n", key)
	} else {
		vPolicy := obj.(*api.Application).DeepCopy()
		glog.Infof("Sync/Add/Update for Application %s/%s\n", vPolicy.Namespace, vPolicy.Name)

		if vPolicy.DeletionTimestamp != nil {
		} else {
			if !core_util.HasFinalizer(vPolicy.ObjectMeta, AppFinalizer) {
				// Add finalizer
				_, _, err := util.PatchApplication(c.extClient.KubepackV1alpha1(), vPolicy, func(vp *api.Application) *api.Application {
					vp.ObjectMeta = core_util.AddFinalizer(vPolicy.ObjectMeta, AppFinalizer)
					return vp
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set Application finalizer for %s/%s", vPolicy.Namespace, vPolicy.Name)
				}
			}

			err = c.reconcilePolicy(vPolicy)
			if err != nil {
				return errors.Wrapf(err, "for Application %s/%s", vPolicy.Namespace, vPolicy.Name)
			}
		}
	}
	return nil
}

func (c *KubepackController) reconcilePolicy(vPolicy *api.Application) error {
	return nil
}
