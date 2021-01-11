package lib

import (
	a2 "kubepack.dev/cli/pkg/lib/action"
	"kubepack.dev/lib-helm/action"

	"gopkg.in/macaron.v1"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"kmodules.xyz/resource-metadata/hub"
)

func ApplyResource(ctx *macaron.Context, model EditorModel, f cmdutil.Factory) (*release.Release, error) {
	opts := EditorOptions{
		Group:       ctx.Params(":group"),
		Version:     ctx.Params(":version"),
		Resource:    ctx.Params(":resource"),
		ReleaseName: ctx.Params(":releaseName"),
		Namespace:   ctx.Params(":namespace"),
		// ValuesFile:  params.ValuesFile,
		// ValuesPatch: params.ValuesPatch,
	}
	rd, err := hub.NewRegistryOfKnownResources().LoadByGVR(schema.GroupVersionResource{
		Group:    opts.Group,
		Version:  opts.Version,
		Resource: opts.Resource,
	})
	if err != nil {
		return nil, err
	}

	applier, err := a2.NewApplier(f, opts.Namespace, "applications")
	if err != nil {
		return nil, err
	}

	applier.WithRegistry(DefaultRegistry)
	opts2 := a2.NewApplyOptions()
	opts2.ChartURL = rd.Spec.UI.Editor.URL
	opts2.ChartName = rd.Spec.UI.Editor.Name
	opts2.Version = rd.Spec.UI.Editor.Version
	opts2.Values = model.Values
	//opts2.ValuesFile =               "values.yaml"
	//opts2.ValuesPatch =              nil
	opts2.CreateNamespace = false // TODO?
	opts2.DryRun = false
	opts2.DisableHooks = false
	opts2.Replace = false
	opts2.Wait = false
	opts2.Timeout = 0
	opts2.Description = "Apply editor"
	opts2.Devel = false
	opts2.Namespace = opts.Namespace
	opts2.ReleaseName = opts.ReleaseName
	opts2.Atomic = false
	opts2.SkipCRDs = false
	opts2.SubNotes = false
	opts2.DisableOpenAPIValidation = false
	opts2.IncludeCRDs = false
	applier.WithOptions(opts2)

	return applier.Run()
}

func DeleteResource(ctx *macaron.Context, f cmdutil.Factory) (*release.UninstallReleaseResponse, error) {
	opts := EditorOptions{
		Group:       ctx.Params(":group"),
		Version:     ctx.Params(":version"),
		Resource:    ctx.Params(":resource"),
		ReleaseName: ctx.Params(":releaseName"),
		Namespace:   ctx.Params(":namespace"),
		// ValuesFile:  params.ValuesFile,
		// ValuesPatch: params.ValuesPatch,
	}

	cmd, err := action.NewUninstaller(f, opts.Namespace, "applications")
	if err != nil {
		return nil, err
	}

	cmd.WithReleaseName(opts.ReleaseName)
	cmd.WithOptions(action.UninstallOptions{
		DisableHooks: false,
		DryRun:       false,
		KeepHistory:  false,
		Timeout:      0,
	})
	return cmd.Run()
}
