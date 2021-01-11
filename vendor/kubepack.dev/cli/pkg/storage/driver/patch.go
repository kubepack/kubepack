package driver

import (
	"context"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kutil "kmodules.xyz/client-go"
	api "sigs.k8s.io/application/api/app/v1beta1"
	cs "sigs.k8s.io/application/client/clientset/versioned/typed/app/v1beta1"
)

func createOrPatchApplication(
	ctx context.Context,
	c cs.ApplicationInterface,
	meta metav1.ObjectMeta,
	transform func(*api.Application) *api.Application,
	opts metav1.PatchOptions,
) (*api.Application, kutil.VerbType, error) {
	cur, err := c.Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Application %s/%s.", meta.Namespace, meta.Name)
		out, err := c.Create(ctx, transform(&api.Application{
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
	return patchApplication(ctx, c, cur, transform, opts)
}

func patchApplication(
	ctx context.Context,
	c cs.ApplicationInterface,
	cur *api.Application,
	transform func(*api.Application) *api.Application,
	opts metav1.PatchOptions,
) (*api.Application, kutil.VerbType, error) {
	return patchApplicationObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func patchApplicationObject(
	ctx context.Context,
	c cs.ApplicationInterface,
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
	out, err := c.Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}
