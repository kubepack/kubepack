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
	"context"
	"encoding/json"
	"fmt"

	api "kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	cs "kubepack.dev/kubepack/client/clientset/versioned/typed/kubepack/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchOrder(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(in *api.Order) *api.Order,
	opts metav1.PatchOptions,
) (*api.Order, kutil.VerbType, error) {
	cur, err := c.Orders().Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Order %s/%s.", meta.Namespace, meta.Name)
		out, err := c.Orders().Create(ctx, transform(&api.Order{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Order",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchOrder(ctx, c, cur, transform, opts)
}

func PatchOrder(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	cur *api.Order,
	transform func(*api.Order) *api.Order,
	opts metav1.PatchOptions,
) (*api.Order, kutil.VerbType, error) {
	return PatchOrderObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchOrderObject(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	cur, mod *api.Order,
	opts metav1.PatchOptions,
) (*api.Order, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching Order %s with %s.", cur.Name, string(patch))
	out, err := c.Orders().Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateOrder(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.Order) *api.Order,
	opts metav1.UpdateOptions,
) (result *api.Order, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Orders().Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.Orders().Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Order %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update Order %s after %d attempts due to %v", meta.Name, attempt, err)
	}
	return
}
