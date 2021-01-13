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

package handler

import (
	"gopkg.in/macaron.v1"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"kmodules.xyz/resource-metadata/hub"
	"kubepack.dev/cli/pkg/lib/action"
	"kubepack.dev/kubepack/pkg/lib"
)

func ApplyResource(ctx *macaron.Context, model unstructured.Unstructured, f cmdutil.Factory) (*release.Release, error) {
	gvr := schema.GroupVersionResource{
		Group:    ctx.Params(":group"),
		Version:  ctx.Params(":version"),
		Resource: ctx.Params(":resource"),
	}
	rlm := lib.ReleaseMetadata{
		Name:      ctx.Params(":releaseName"),
		Namespace: ctx.Params(":namespace"),
	}
	rd, err := hub.NewRegistryOfKnownResources().LoadByGVR(gvr)
	if err != nil {
		return nil, err
	}

	applier, err := action.NewApplier(f, rlm.Namespace, "applications")
	if err != nil {
		return nil, err
	}

	applier.WithRegistry(lib.DefaultRegistry)
	opts2 := action.NewApplyOptions()
	opts2.ChartURL = rd.Spec.UI.Editor.URL
	opts2.ChartName = rd.Spec.UI.Editor.Name
	opts2.Version = rd.Spec.UI.Editor.Version
	opts2.Values = model.Object
	//opts2.ValuesFile =               "values.yaml"
	//opts2.ValuesPatch =              nil
	opts2.CreateNamespace = true // TODO?
	opts2.DryRun = false
	opts2.DisableHooks = false
	opts2.Replace = false
	opts2.Wait = false
	opts2.Timeout = 0
	opts2.Description = "Apply editor"
	opts2.Devel = false
	opts2.Namespace = rlm.Namespace
	opts2.ReleaseName = rlm.Name
	opts2.Atomic = false
	opts2.SkipCRDs = false
	opts2.SubNotes = false
	opts2.DisableOpenAPIValidation = false
	opts2.IncludeCRDs = false
	applier.WithOptions(opts2)

	return applier.Run()
}

func DeleteResource(ctx *macaron.Context, f cmdutil.Factory) (*release.UninstallReleaseResponse, error) {
	rlm := lib.ReleaseMetadata{
		Name:      ctx.Params(":releaseName"),
		Namespace: ctx.Params(":namespace"),
	}

	cmd, err := action.NewUninstaller(f, rlm.Namespace, "applications")
	if err != nil {
		return nil, err
	}

	cmd.WithReleaseName(rlm.Name)
	cmd.WithOptions(action.UninstallOptions{
		DisableHooks: false,
		DryRun:       false,
		KeepHistory:  false,
		Timeout:      0,
	})
	return cmd.Run()
}
