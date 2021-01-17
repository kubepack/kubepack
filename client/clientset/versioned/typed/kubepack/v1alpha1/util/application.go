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

package util

import (
	"context"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
	api "sigs.k8s.io/application/api/app/v1beta1"
	cs "sigs.k8s.io/application/client/clientset/versioned/typed/app/v1beta1"
)

func CreateOrPatchApplication(
	ctx context.Context,
	c cs.ApplicationsGetter,
	meta metav1.ObjectMeta,
	transform func(*api.Application) *api.Application,
	opts metav1.PatchOptions,
) (*api.Application, kutil.VerbType, error) {
	cur, err := c.Applications(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Application %s/%s.", meta.Namespace, meta.Name)
		out, err := c.Applications(meta.Namespace).Create(ctx, transform(&api.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: api.GroupVersion.String(),
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
	return PatchApplication(ctx, c, cur, transform, opts)
}

func PatchApplication(
	ctx context.Context,
	c cs.ApplicationsGetter,
	cur *api.Application,
	transform func(*api.Application) *api.Application,
	opts metav1.PatchOptions,
) (*api.Application, kutil.VerbType, error) {
	return PatchApplicationObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchApplicationObject(
	ctx context.Context,
	c cs.ApplicationsGetter,
	cur, mod *api.Application,
	opts metav1.PatchOptions,
) (*api.Application, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching Application %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.Applications(cur.Namespace).Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateApplication(
	ctx context.Context,
	c cs.ApplicationsGetter,
	meta metav1.ObjectMeta,
	transform func(*api.Application) *api.Application,
	opts metav1.UpdateOptions,
) (result *api.Application, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Applications(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.Applications(cur.Namespace).Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Application %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update Application %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateApplicationStatus(
	ctx context.Context,
	c cs.ApplicationsGetter,
	meta metav1.ObjectMeta,
	transform func(*api.ApplicationStatus) *api.ApplicationStatus,
	opts metav1.UpdateOptions,
) (result *api.Application, err error) {
	apply := func(x *api.Application) *api.Application {
		return &api.Application{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(x.Status.DeepCopy()),
		}
	}

	attempt := 0
	cur, err := c.Applications(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.Applications(meta.Namespace).UpdateStatus(ctx, apply(cur), opts)
		if kerr.IsConflict(e2) {
			latest, e3 := c.Applications(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
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
		err = fmt.Errorf("failed to update status of Application %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
