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

package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/lib-app/api"
	"kubepack.dev/lib-helm/repo"

	jsonpatch "github.com/evanphx/json-patch"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"gomodules.xyz/version"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime"
	yamllib "sigs.k8s.io/yaml"
)

type TemplateRenderer struct {
	Registry    *repo.Registry
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	KubeVersion string
	ValuesFile  string
	ValuesPatch *runtime.RawExtension
	Values      map[string]interface{}

	BucketURL string
	UID       string
	PublicURL string
	//W         io.Writer

	CRDs     []api.BucketFile
	Manifest *api.BucketFile
}

func (x *TemplateRenderer) Do() error {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, x.BucketURL)
	if err != nil {
		return err
	}

	dirManifest := blob.PrefixedBucket(bucket, x.UID+"/manifests/")
	defer dirManifest.Close()
	dirCRD := blob.PrefixedBucket(bucket, x.UID+"/crds/")
	defer dirCRD.Close()

	chrt, err := x.Registry.GetChart(x.ChartRef.URL, x.ChartRef.Name, x.Version)
	if err != nil {
		return err
	}

	cfg := new(action.Configuration)
	client := action.NewInstall(cfg)
	var extraAPIs []string

	client.DryRun = true
	client.ReleaseName = x.ReleaseName
	client.Namespace = x.Namespace
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.APIVersions = chartutil.VersionSet(extraAPIs)
	client.Version = x.Version

	validInstallableChart, err := isChartInstallable(chrt.Chart)
	if !validInstallableChart {
		return err
	}

	//if chrt.Metadata.Deprecated {
	//}

	if req := chrt.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chrt.Chart, req); err != nil {
			return err
		}
	}

	vals := chrt.Values
	if x.ValuesPatch != nil {
		if x.ValuesFile != "" {
			for _, f := range chrt.Raw {
				if f.Name == x.ValuesFile {
					if err := yamllib.Unmarshal(f.Data, &vals); err != nil {
						return fmt.Errorf("cannot load %s. Reason: %v", f.Name, err.Error())
					}
					break
				}
			}
		}
		values, err := json.Marshal(vals)
		if err != nil {
			return err
		}

		patchData, err := json.Marshal(x.ValuesPatch)
		if err != nil {
			return err
		}
		patch, err := jsonpatch.DecodePatch(patchData)
		if err != nil {
			return err
		}
		modifiedValues, err := patch.Apply(values)
		if err != nil {
			return err
		}
		err = json.Unmarshal(modifiedValues, &vals)
		if err != nil {
			return err
		}
	} else if x.Values != nil {
		vals = x.Values
	}

	// Pre-install anything in the crd/ directory. We do this before Helm
	// contacts the upstream server and builds the capabilities object.
	if crds := chrt.CRDObjects(); len(crds) > 0 {
		for _, crd := range crds {
			// Open the key "${releaseName}.yaml" for writing with the default options.
			w, err := dirCRD.NewWriter(ctx, crd.Name+".yaml", nil)
			if err != nil {
				return err
			}
			_, writeErr := w.Write(crd.File.Data)
			// Always check the return value of Close when writing.
			closeErr := w.Close()
			if writeErr != nil {
				return writeErr
			}
			if closeErr != nil {
				return closeErr
			}

			objectKey := "/" + path.Join(x.UID, "crds", crd.Name+".yaml")
			x.CRDs = append(x.CRDs, api.BucketFile{
				URL:      x.PublicURL + objectKey,
				Key:      objectKey,
				Filename: crd.Filename,
				Data:     crd.File.Data,
			})
		}
	}

	if err := chartutil.ProcessDependencies(chrt.Chart, vals); err != nil {
		return err
	}

	caps := chartutil.DefaultCapabilities
	if x.KubeVersion != "" {
		info, err := version.NewSemver(x.KubeVersion)
		if err != nil {
			return err
		}
		info = info.ToMutator().ResetPrerelease().ResetMetadata().Done()
		caps.KubeVersion = chartutil.KubeVersion{
			Version: info.String(),
			Major:   strconv.FormatInt(info.Major(), 10),
			Minor:   strconv.FormatInt(info.Minor(), 10),
		}
	}
	options := chartutil.ReleaseOptions{
		Name:      x.ReleaseName,
		Namespace: x.Namespace,
		Revision:  1,
		IsInstall: true,
	}
	valuesToRender, err := chartutil.ToRenderValues(chrt.Chart, vals, options, caps)
	if err != nil {
		return err
	}
	if x.Values != nil {
		valuesToRender["Values"] = x.Values
	}

	hooks, manifests, err := renderResources(chrt.Chart, caps, valuesToRender)
	if err != nil {
		return err
	}

	var manifestDoc bytes.Buffer

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPreInstall) {
			// TODO: Mark as pre-install hook
			_, err = fmt.Fprintf(&manifestDoc, "---\n# Source: %s\n%s\n", hook.Path, hook.Manifest)
			if err != nil {
				return err
			}
		}
	}

	for _, m := range manifests {
		_, err = fmt.Fprintf(&manifestDoc, "---\n# Source: %s\n%s\n", m.Name, m.Content)
		if err != nil {
			return err
		}
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPostInstall) {
			// TODO: Mark as post-install hook
			_, err = fmt.Fprintf(&manifestDoc, "---\n# Source: %s\n%s\n", hook.Path, hook.Manifest)
			if err != nil {
				return err
			}
		}
	}

	{
		objectKey := "/" + path.Join(x.UID, "manifests", x.ReleaseName+".yaml")
		x.Manifest = &api.BucketFile{
			URL:      x.PublicURL + objectKey,
			Key:      objectKey,
			Filename: "manifest.yaml",
			Data:     manifestDoc.Bytes(),
		}

		// Open the key "${releaseName}.yaml" for writing with the default options.
		w, err := dirManifest.NewWriter(ctx, x.ReleaseName+".yaml", nil)
		if err != nil {
			return err
		}
		_, writeErr := manifestDoc.WriteTo(w)
		// Always check the return value of Close when writing.
		closeErr := w.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}

	return nil
}

func (x *TemplateRenderer) Result() (crds []api.BucketFile, manifest *api.BucketFile) {
	crds = x.CRDs
	manifest = x.Manifest

	return
}

type EditorModelGenerator struct {
	Registry    *repo.Registry
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	KubeVersion string
	ValuesFile  string
	ValuesPatch *runtime.RawExtension
	Values      map[string]interface{}

	CRDs     []*chart.File
	Manifest []byte
}

func (x *EditorModelGenerator) Do() error {
	chrt, err := x.Registry.GetChart(x.ChartRef.URL, x.ChartRef.Name, x.Version)
	if err != nil {
		return err
	}

	cfg := new(action.Configuration)
	client := action.NewInstall(cfg)
	var extraAPIs []string

	client.DryRun = true
	client.ReleaseName = x.ReleaseName
	client.Namespace = x.Namespace
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.APIVersions = chartutil.VersionSet(extraAPIs)
	client.Version = x.Version

	validInstallableChart, err := isChartInstallable(chrt.Chart)
	if !validInstallableChart {
		return err
	}

	//if chrt.Metadata.Deprecated {
	//}

	if req := chrt.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chrt.Chart, req); err != nil {
			return err
		}
	}

	vals := chrt.Values
	if x.ValuesPatch != nil {
		if x.ValuesFile != "" {
			for _, f := range chrt.Raw {
				if f.Name == x.ValuesFile {
					if err := yamllib.Unmarshal(f.Data, &vals); err != nil {
						return fmt.Errorf("cannot load %s. Reason: %v", f.Name, err.Error())
					}
					break
				}
			}
		}
		values, err := json.Marshal(vals)
		if err != nil {
			return err
		}

		patchData, err := json.Marshal(x.ValuesPatch)
		if err != nil {
			return err
		}
		patch, err := jsonpatch.DecodePatch(patchData)
		if err != nil {
			return err
		}
		modifiedValues, err := patch.Apply(values)
		if err != nil {
			return err
		}
		err = json.Unmarshal(modifiedValues, &vals)
		if err != nil {
			return err
		}
	} else if x.Values != nil {
		vals = x.Values
	}

	// Pre-install anything in the crd/ directory. We do this before Helm
	// contacts the upstream server and builds the capabilities object.
	for _, crd := range chrt.CRDObjects() {
		x.CRDs = append(x.CRDs, crd.File)
	}

	if err := chartutil.ProcessDependencies(chrt.Chart, vals); err != nil {
		return err
	}

	caps := chartutil.DefaultCapabilities
	if x.KubeVersion != "" {
		info, err := version.NewSemver(x.KubeVersion)
		if err != nil {
			return err
		}
		info = info.ToMutator().ResetPrerelease().ResetMetadata().Done()
		caps.KubeVersion = chartutil.KubeVersion{
			Version: info.String(),
			Major:   strconv.FormatInt(info.Major(), 10),
			Minor:   strconv.FormatInt(info.Minor(), 10),
		}
	}
	options := chartutil.ReleaseOptions{
		Name:      x.ReleaseName,
		Namespace: x.Namespace,
		Revision:  1,
		IsInstall: true,
	}
	valuesToRender, err := chartutil.ToRenderValues(chrt.Chart, vals, options, caps)
	if err != nil {
		return err
	}
	if x.Values != nil {
		valuesToRender["Values"] = x.Values
	}

	hooks, manifests, err := renderResources(chrt.Chart, caps, valuesToRender)
	if err != nil {
		return err
	}

	var manifestDoc bytes.Buffer

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPreInstall) {
			// TODO: Mark as pre-install hook
			_, err = fmt.Fprintf(&manifestDoc, "---\n# Source: %s\n%s\n", hook.Path, hook.Manifest)
			if err != nil {
				return err
			}
		}
	}

	for _, m := range manifests {
		_, err = fmt.Fprintf(&manifestDoc, "---\n# Source: %s\n%s\n", m.Name, m.Content)
		if err != nil {
			return err
		}
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPostInstall) {
			// TODO: Mark as post-install hook
			_, err = fmt.Fprintf(&manifestDoc, "---\n# Source: %s\n%s\n", hook.Path, hook.Manifest)
			if err != nil {
				return err
			}
		}
	}

	{
		x.Manifest = manifestDoc.Bytes()
	}

	return nil
}

func (x *EditorModelGenerator) Result() ([]*chart.File, []byte) {
	return x.CRDs, x.Manifest
}
