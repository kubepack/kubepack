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
	"encoding/json"
	"fmt"

	api "kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	cs "kubepack.dev/kubepack/client/clientset/versioned/typed/kubepack/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchOrder(c cs.KubepackV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.Order) *api.Order) (*api.Order, kutil.VerbType, error) {
	cur, err := c.Orders(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Order %s/%s.", meta.Namespace, meta.Name)
		out, err := c.Orders(meta.Namespace).Create(transform(&api.Order{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindOrder,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchOrder(c, cur, transform)
}

func PatchOrder(c cs.KubepackV1alpha1Interface, cur *api.Order, transform func(*api.Order) *api.Order) (*api.Order, kutil.VerbType, error) {
	return PatchOrderObject(c, cur, transform(cur.DeepCopy()))
}

func PatchOrderObject(c cs.KubepackV1alpha1Interface, cur, mod *api.Order) (*api.Order, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching Order %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.Orders(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateOrder(c cs.KubepackV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.Order) *api.Order) (result *api.Order, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Orders(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.Orders(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Order %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update Order %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateOrderStatus(
	c cs.KubepackV1alpha1Interface,
	in *api.Order,
	transform func(*api.OrderStatus) *api.OrderStatus,
) (result *api.Order, err error) {
	apply := func(x *api.Order, copy bool) *api.Order {
		out := &api.Order{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
		}
		if copy {
			out.Status = *transform(in.Status.DeepCopy())
		} else {
			out.Status = *transform(&in.Status)
		}
		return out
	}

	attempt := 0
	cur := in.DeepCopy()
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.Orders(in.Namespace).UpdateStatus(apply(cur, false))
		if kerr.IsConflict(e2) {
			latest, e3 := c.Orders(in.Namespace).Get(in.Name, metav1.GetOptions{})
			switch {
			case e3 == nil:
				cur = latest
				return false, nil
			case kutil.IsRequestRetryable(e3):
				return false, nil
			default:
				return false, e3
			}
		} else if err != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return e2 == nil, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update status of Order %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
	}
	return
}
