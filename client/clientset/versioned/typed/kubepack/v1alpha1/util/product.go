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
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchProduct(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.Product) *api.Product,
	opts metav1.PatchOptions,
) (*api.Product, kutil.VerbType, error) {
	cur, err := c.Products().Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Product %s.", meta.Name)
		out, err := c.Products().Create(ctx, transform(&api.Product{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindProduct,
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
	return PatchProduct(ctx, c, cur, transform, opts)
}

func PatchProduct(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	cur *api.Product,
	transform func(*api.Product) *api.Product,
	opts metav1.PatchOptions,
) (*api.Product, kutil.VerbType, error) {
	return PatchProductObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchProductObject(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	cur, mod *api.Product,
	opts metav1.PatchOptions,
) (*api.Product, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching Product %s with %s.", cur.Name, string(patch))
	out, err := c.Products().Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateProduct(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.Product) *api.Product,
	opts metav1.UpdateOptions,
) (result *api.Product, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Products().Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.Products().Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Product %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update Product %s after %d attempts due to %v", meta.Name, attempt, err)
	}
	return
}

func UpdateProductStatus(
	ctx context.Context,
	c cs.KubepackV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.ProductStatus) *api.ProductStatus,
	opts metav1.UpdateOptions,
) (result *api.Product, err error) {
	apply := func(x *api.Product) *api.Product {
		return &api.Product{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(x.Status.DeepCopy()),
		}
	}

	attempt := 0
	cur, err := c.Products().Get(ctx, meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.Products().UpdateStatus(ctx, apply(cur), opts)
		if kerr.IsConflict(e2) {
			latest, e3 := c.Products().Get(ctx, meta.Name, metav1.GetOptions{})
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
		err = fmt.Errorf("failed to update status of Product %s after %d attempts due to %v", meta.Name, attempt, err)
	}
	return
}
