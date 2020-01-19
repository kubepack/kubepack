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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"golang.org/x/net/publicsuffix"
	"gomodules.xyz/version"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	authorization "k8s.io/api/authorization/v1"
	core "k8s.io/api/core/v1"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"kmodules.xyz/client-go/apiextensions/v1beta1"
	"kmodules.xyz/lib-chart/helm"
	"kmodules.xyz/resource-metadata/hub"
	yamllib "sigs.k8s.io/yaml"
)

type DoFn func() error

type NamespacePrinter struct {
	Namespace string
	W         io.Writer
}

func (x *NamespacePrinter) Do() error {
	if x.Namespace == core.NamespaceDefault || x.Namespace == "kube-system" {
		return nil
	}

	_, err := x.W.Write([]byte("## create namespace if missing\n"))
	if err != nil {
		return err
	}
	_, err = x.W.Write([]byte(fmt.Sprintf("kubectl create namespace %s || true\n", x.Namespace)))
	return err
}

type NamespaceCreator struct {
	Namespace string
	Client    kubernetes.Interface
}

func (x *NamespaceCreator) Do() error {
	if x.Namespace == core.NamespaceDefault || x.Namespace == "kube-system" {
		return nil
	}

	_, err := x.Client.CoreV1().Namespaces().Create(&core.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: x.Namespace,
		},
	})
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}

type WaitForPrinter struct {
	Name      string
	Namespace string
	WaitFors  []v1alpha1.WaitOptions
	W         io.Writer
}

func (x *WaitForPrinter) Do() error {
	if len(x.WaitFors) == 0 {
		return nil
	}

	_, err := fmt.Fprintf(x.W, "## wait %s to be ready\n", x.Name)
	if err != nil {
		return err
	}
	for _, w := range x.WaitFors {
		// kubectl wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available] [options]

		parts := make([]string, 0, 10)
		parts = append(parts, "kubectl")
		parts = append(parts, "wait")

		if w.Resource.Group != "" {
			if w.Resource.Name != "" {
				parts = append(parts, w.Resource.Group+"/"+w.Resource.Name)
			} else {
				parts = append(parts, w.Resource.Group)
			}
		}

		if w.Labels != nil {
			parts = append(parts, "-l")

			selector, err := v1.LabelSelectorAsSelector(w.Labels)
			if err != nil {
				return err
			}
			parts = append(parts, selector.String())
		}

		if w.All {
			parts = append(parts, "--all")
		}

		if w.ForCondition != "" {
			parts = append(parts, "--for")
			parts = append(parts, w.ForCondition)
		}

		if w.Timeout.Duration > 0 {
			parts = append(parts, "--timeout")
			parts = append(parts, w.Timeout.Duration.String())
		}

		if x.Namespace != "" {
			parts = append(parts, "-n")
			parts = append(parts, x.Namespace)
		}

		_, err = fmt.Fprintln(x.W, strings.Join(parts, " "))
		if err != nil {
			return err
		}
	}
	return nil
}

type WaitForChecker struct {
	Name      string
	Namespace string
	WaitFors  []v1alpha1.WaitOptions
	W         io.Writer
}

func (x *WaitForChecker) Do() error {
	return nil
}

type CRDReadinessPrinter struct {
	CRDs []v1alpha1.ResourceID
	W    io.Writer
}

func (x *CRDReadinessPrinter) Do() error {
	_, err := fmt.Fprintln(x.W, "## wait for crds to be ready")
	if err != nil {
		return err
	}

	for _, crd := range x.CRDs {
		_, err := fmt.Fprintf(x.W, "kubectl wait --for=condition=Established crds/%s.%s --timeout=5m\n", crd.Name, crd.Group)
		if err != nil {
			return err
		}
	}
	return nil
}

type CRDReadinessChecker struct {
	CRDs   []v1alpha1.ResourceID
	Client rest.Interface
}

func (x *CRDReadinessChecker) Do() error {
	crds := make([]*crdv1beta1.CustomResourceDefinition, 0, len(x.CRDs))
	for _, crd := range x.CRDs {
		crds = append(crds, &crdv1beta1.CustomResourceDefinition{
			ObjectMeta: v1.ObjectMeta{
				Name: fmt.Sprintf("%s.%s", crd.Name, crd.Group),
			},
			Spec: crdv1beta1.CustomResourceDefinitionSpec{
				Group:   crd.Group,
				Version: crd.Version,
				Names: crdv1beta1.CustomResourceDefinitionNames{
					Plural: crd.Name,
					Kind:   crd.Kind,
				},
				Scope: crdv1beta1.ResourceScope(string(crd.Scope)),
				Versions: []crdv1beta1.CustomResourceDefinitionVersion{
					{
						Name: crd.Version,
					},
				},
			},
		})
	}
	return v1beta1.WaitForCRDReady(x.Client, crds)
}

type Helm3CommandPrinter struct {
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	ValuesPatch *runtime.RawExtension

	W io.Writer
}

const indent = "  "

func (x *Helm3CommandPrinter) Do() error {
	chrt, err := GetChart(x.ChartRef.Name, x.Version, "myrepo", x.ChartRef.URL)
	if err != nil {
		return err
	}

	reponame, err := RepoName(x.ChartRef.URL)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	/*
		$ helm repo add appscode https://charts.appscode.com/stable/
		$ helm repo update
		$ helm search repo appscode/voyager --version v12.0.0-rc.1
	*/
	_, err = fmt.Fprintf(&buf, "## add helm repository %s\n", reponame)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm repo add %s %s\n", reponame, x.ChartRef.URL)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm repo update\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm search repo %s/%s --version %s\n", reponame, x.ChartRef.Name, x.Version)
	if err != nil {
		return err
	}

	/*
		$ helm install voyager-operator appscode/voyager --version v12.0.0-rc.1 \
		  --namespace kube-system \
		  --set cloudProvider=$provider
	*/
	_, err = fmt.Fprintf(&buf, "## install chart %s/%s\n", reponame, x.ChartRef.Name)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm install %s %s/%s --version %s \\\n", x.ReleaseName, reponame, x.ChartRef.Name, x.Version)
	if err != nil {
		return err
	}
	if x.Namespace != "" {
		_, err = fmt.Fprintf(&buf, "%s--namespace %s \\\n", indent, x.Namespace)
		if err != nil {
			return err
		}
	}

	if x.ValuesPatch != nil {
		values, err := json.Marshal(chrt.Values)
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
		var modified map[string]interface{}
		err = json.Unmarshal(modifiedValues, &modified)
		if err != nil {
			return err
		}
		setValues := helm.GetChangedValues(chrt.Values, modified)
		for _, v := range setValues {
			_, err = fmt.Fprintf(&buf, `%s--set %s \\n`, indent, v)
			if err != nil {
				return err
			}
		}
	}
	buf.Truncate(buf.Len() - 3)

	_, err = buf.WriteRune('\n')
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(x.W)
	return err
}

type Helm2CommandPrinter struct {
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	ValuesPatch *runtime.RawExtension

	W io.Writer
}

func (x *Helm2CommandPrinter) Do() error {
	chrt, err := GetChart(x.ChartRef.Name, x.Version, "myrepo", x.ChartRef.URL)
	if err != nil {
		return err
	}

	reponame, err := RepoName(x.ChartRef.URL)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	/*
		$ helm repo add appscode https://charts.appscode.com/stable/
		$ helm repo update
		$ helm search repo appscode/voyager --version v12.0.0-rc.1
	*/
	_, err = fmt.Fprintf(&buf, "## add helm repository %s\n", reponame)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm repo add %s %s\n", reponame, x.ChartRef.URL)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm repo update\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm search %s/%s --version %s\n", reponame, x.ChartRef.Name, x.Version)
	if err != nil {
		return err
	}

	/*
		$ helm install voyager-operator appscode/voyager --version v12.0.0-rc.1 \
		  --namespace kube-system \
		  --set cloudProvider=$provider
	*/
	_, err = fmt.Fprintf(&buf, "## install chart %s/%s\n", reponame, x.ChartRef.Name)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(&buf, "helm install %s/%s --name %s --version %s \\\n", reponame, x.ChartRef.Name, x.ReleaseName, x.Version)
	if err != nil {
		return err
	}
	if x.Namespace != "" {
		_, err = fmt.Fprintf(&buf, "%s--namespace %s \\\n", indent, x.Namespace)
		if err != nil {
			return err
		}
	}

	if x.ValuesPatch != nil {
		values, err := json.Marshal(chrt.Values)
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
		var modified map[string]interface{}
		err = json.Unmarshal(modifiedValues, &modified)
		if err != nil {
			return err
		}
		setValues := helm.GetChangedValues(chrt.Values, modified)
		for _, v := range setValues {
			_, err = fmt.Fprintf(&buf, `%s--set %s \\n`, indent, v)
			if err != nil {
				return err
			}
		}
	}
	buf.Truncate(buf.Len() - 3)

	_, err = buf.WriteRune('\n')
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(x.W)
	return err
}

const YAMLHost = "https://usercontent.kubepack.com"
const YAMLBucket = "gs://kubepack-usercontent"

type YAMLPrinter struct {
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	KubeVersion string
	ValuesPatch *runtime.RawExtension

	BucketURL string
	UID       string
	PublicURL string
	W         io.Writer
}

func (x *YAMLPrinter) Do() error {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, x.BucketURL)
	if err != nil {
		return err
	}

	dirManifest := blob.PrefixedBucket(bucket, x.UID+"/manifests/")
	defer dirManifest.Close()
	dirCRD := blob.PrefixedBucket(bucket, x.UID+"/crds/")
	defer dirCRD.Close()

	var buf bytes.Buffer

	chrt, err := GetChart(x.ChartRef.Name, x.Version, "myrepo", x.ChartRef.URL)
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

	validInstallableChart, err := isChartInstallable(chrt)
	if !validInstallableChart {
		return err
	}

	if chrt.Metadata.Deprecated {
		_, err = fmt.Fprintln(&buf, "# WARNING: This chart is deprecated")
		if err != nil {
			return err
		}

	}

	/*
		// cp, err := client.ChartPathOptions.LocateChart(chart, settings)
		//
		if req := chrt.Metadata.Dependencies; req != nil {
			// If CheckDependencies returns an error, we have unfulfilled dependencies.
			// As of Helm 2.4.0, this is treated as a stopping condition:
			// https://github.com/helm/helm/issues/2209
			if err := action.CheckDependencies(chrt, req); err != nil {
				if client.DependencyUpdate {
					man := &downloader.Manager{
						Out:              os.Stdout, // TODO: pass io.Writer?
						ChartPath:        cp,
						Keyring:          client.ChartPathOptions.Keyring,
						SkipUpdate:       false,
						Getters:          p,
						RepositoryConfig: settings.RepositoryConfig,
						RepositoryCache:  settings.RepositoryCache,
					}
					if err := man.Update(); err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	*/

	vals := chrt.Values
	if x.ValuesPatch != nil {
		values, err := json.Marshal(chrt.Values)
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
	}

	// Pre-install anything in the crd/ directory. We do this before Helm
	// contacts the upstream server and builds the capabilities object.
	if crds := chrt.CRDs(); len(crds) > 0 {
		_, err = fmt.Fprintln(&buf, "# install CRDs")
		if err != nil {
			return err
		}

		for _, crd := range crds {
			// Open the key "${releaseName}.yaml" for writing with the default options.
			w, err := dirCRD.NewWriter(ctx, crd.Name+".yaml", nil)
			if err != nil {
				return err
			}
			_, writeErr := w.Write(crd.Data)
			// Always check the return value of Close when writing.
			closeErr := w.Close()
			if writeErr != nil {
				return writeErr
			}
			if closeErr != nil {
				return closeErr
			}

			_, err = fmt.Fprintf(&buf, "kubectl apply -f %s\n", x.PublicURL+"/"+path.Join(x.UID, "crds", crd.Name+".yaml"))
			if err != nil {
				return err
			}
		}
	}

	if err := chartutil.ProcessDependencies(chrt, vals); err != nil {
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
	valuesToRender, err := chartutil.ToRenderValues(chrt, vals, options, caps)
	if err != nil {
		return err
	}

	hooks, manifests, err := renderResources(chrt, caps, valuesToRender)
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

		_, err = fmt.Fprintf(&buf, "kubectl apply -f %s\n", x.PublicURL+"/"+path.Join(x.UID, "manifests", x.ReleaseName+".yaml"))
		if err != nil {
			return err
		}
	}

	_, err = buf.WriteTo(x.W)
	return err
}

type ChartInstaller struct {
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	ValuesPatch *runtime.RawExtension

	ClientGetter genericclioptions.RESTClientGetter
}

func (x *ChartInstaller) Do() error {
	chrt, err := GetChart(x.ChartRef.Name, x.Version, "myrepo", x.ChartRef.URL)
	if err != nil {
		return err
	}

	cfg := new(action.Configuration)
	// TODO: Use secret driver for which namespace?
	err = cfg.Init(x.ClientGetter, x.Namespace, "secret", debug)
	if err != nil {
		return err
	}

	client := action.NewInstall(cfg)
	var extraAPIs []string

	client.DryRun = true
	client.ReleaseName = x.ReleaseName
	client.Namespace = x.Namespace
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.APIVersions = chartutil.VersionSet(extraAPIs)
	client.Version = x.Version

	validInstallableChart, err := isChartInstallable(chrt)
	if !validInstallableChart {
		return err
	}

	if chrt.Metadata.Deprecated {
		_, err = fmt.Println("# WARNING: This chart is deprecated")
		if err != nil {
			return err
		}
	}

	/*
		// cp, err := client.ChartPathOptions.LocateChart(chart, settings)
		//
		if req := chrt.Metadata.Dependencies; req != nil {
			// If CheckDependencies returns an error, we have unfulfilled dependencies.
			// As of Helm 2.4.0, this is treated as a stopping condition:
			// https://github.com/helm/helm/issues/2209
			if err := action.CheckDependencies(chrt, req); err != nil {
				if client.DependencyUpdate {
					man := &downloader.Manager{
						Out:              os.Stdout, // TODO: pass io.Writer?
						ChartPath:        cp,
						Keyring:          client.ChartPathOptions.Keyring,
						SkipUpdate:       false,
						Getters:          p,
						RepositoryConfig: settings.RepositoryConfig,
						RepositoryCache:  settings.RepositoryCache,
					}
					if err := man.Update(); err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	*/

	vals := chrt.Values
	if x.ValuesPatch != nil {
		values, err := json.Marshal(chrt.Values)
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
	}

	_, err = client.Run(chrt, vals)
	if err != nil {
		return err
	}

	// TODO: implement kubectl waits

	return nil
}

type ChartUninstaller struct {
	ReleaseName string
	Namespace   string

	ClientGetter genericclioptions.RESTClientGetter
}

func (x *ChartUninstaller) Do() error {
	cfg := new(action.Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(x.ClientGetter, x.Namespace, "secret", debug)
	if err != nil {
		return err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	client := action.NewUninstall(cfg)
	client.DryRun = false

	_, err = client.Run(x.ReleaseName)
	if err != nil {
		return err
	}

	// TODO: implement kubectl waits

	return nil
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	log.Output(2, fmt.Sprintf(format, v...))
}

type PermissionChecker struct {
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	Verb        string

	Config       *rest.Config
	ClientGetter genericclioptions.RESTClientGetter
	Registry     *hub.Registry

	attrs map[authorization.ResourceAttributes]*ResourcePermission
	m     sync.Mutex
}

type ResourcePermission struct {
	Items   []*unstructured.Unstructured
	Allowed bool
}

func (x *PermissionChecker) Do() error {
	if x.attrs == nil {
		x.attrs = make(map[authorization.ResourceAttributes]*ResourcePermission)
	}

	chrt, err := GetChart(x.ChartRef.Name, x.Version, "myrepo", x.ChartRef.URL)
	if err != nil {
		return err
	}

	cfg := new(action.Configuration)
	err = cfg.Init(x.ClientGetter, x.Namespace, "memory", debug)
	if err != nil {
		return err
	}

	client := action.NewInstall(cfg)
	var extraAPIs []string

	client.DryRun = true
	client.ReleaseName = x.ReleaseName
	client.Namespace = x.Namespace
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.APIVersions = chartutil.VersionSet(extraAPIs)
	client.Version = x.Version

	validInstallableChart, err := isChartInstallable(chrt)
	if !validInstallableChart {
		return err
	}

	//if chrt.Metadata.Deprecated {
	//	_, err = fmt.Fprintln(&buf, "# WARNING: This chart is deprecated")
	//	if err != nil {
	//		return err
	//	}
	//
	//}

	/*
		// cp, err := client.ChartPathOptions.LocateChart(chart, settings)
		//
		if req := chrt.Metadata.Dependencies; req != nil {
			// If CheckDependencies returns an error, we have unfulfilled dependencies.
			// As of Helm 2.4.0, this is treated as a stopping condition:
			// https://github.com/helm/helm/issues/2209
			if err := action.CheckDependencies(chrt, req); err != nil {
				if client.DependencyUpdate {
					man := &downloader.Manager{
						Out:              os.Stdout, // TODO: pass io.Writer?
						ChartPath:        cp,
						Keyring:          client.ChartPathOptions.Keyring,
						SkipUpdate:       false,
						Getters:          p,
						RepositoryConfig: settings.RepositoryConfig,
						RepositoryCache:  settings.RepositoryCache,
					}
					if err := man.Update(); err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	*/

	vals := chrt.Values
	//if x.ValuesPatch != nil {
	//	values, err := json.Marshal(chrt.Values)
	//	if err != nil {
	//		return err
	//	}
	//
	//	patchData, err := json.Marshal(x.ValuesPatch)
	//	if err != nil {
	//		return err
	//	}
	//	patch, err := jsonpatch.DecodePatch(patchData)
	//	if err != nil {
	//		return err
	//	}
	//	modifiedValues, err := patch.Apply(values)
	//	if err != nil {
	//		return err
	//	}
	//	err = json.Unmarshal(modifiedValues, &vals)
	//	if err != nil {
	//		return err
	//	}
	//}

	// Pre-install anything in the crd/ directory. We do this before Helm
	// contacts the upstream server and builds the capabilities object.
	if crds := chrt.CRDs(); len(crds) > 0 {
		attr := authorization.ResourceAttributes{
			Verb:     x.Verb,
			Group:    "apiextensions.k8s.io",
			Version:  "v1beta1",
			Resource: "CustomResourceDefinition",
		}
		info, found := x.attrs[attr]
		if !found {
			info = new(ResourcePermission)
			x.attrs[attr] = info
		}

		for _, crd := range crds {
			items, err := ExtractResources(crd.Data)
			if err != nil {
				return err
			}
			info.Items = append(info.Items, items...)
		}
	}

	if err := chartutil.ProcessDependencies(chrt, vals); err != nil {
		return err
	}

	caps, err := cfg.GetCapabilities()
	if err != nil {
		return err
	}
	options := chartutil.ReleaseOptions{
		Name:      x.ReleaseName,
		Namespace: x.Namespace,
		Revision:  1,
		IsInstall: true,
	}
	valuesToRender, err := chartutil.ToRenderValues(chrt, vals, options, caps)
	if err != nil {
		return err
	}

	hooks, manifests, err := renderResources(chrt, caps, valuesToRender)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPreInstall) {
			err = ExtractResourceAttributes([]byte(hook.Manifest), x.Verb, x.Registry, x.attrs)
			if err != nil {
				return err
			}
		}
	}

	for _, m := range manifests {
		err = ExtractResourceAttributes([]byte(m.Content), x.Verb, x.Registry, x.attrs)
		if err != nil {
			return err
		}
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPostInstall) {
			err = ExtractResourceAttributes([]byte(hook.Manifest), x.Verb, x.Registry, x.attrs)
			if err != nil {
				return err
			}
		}
	}

	ac, err := authv1client.NewForConfig(x.Config)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for attr := range x.attrs {
		wg.Add(1)
		go func(attr authorization.ResourceAttributes) {
			defer wg.Done()

			result, err := ac.SelfSubjectAccessReviews().Create(&authorization.SelfSubjectAccessReview{
				Spec: authorization.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &attr,
				},
			})
			if err != nil {
				panic(err) // TODO: return err
			}

			x.m.Lock()
			x.attrs[attr].Allowed = result.Status.Allowed
			x.m.Unlock()
		}(attr)
	}
	wg.Wait()

	return nil
}

func (x *PermissionChecker) Result() (map[authorization.ResourceAttributes]*ResourcePermission, bool) {
	for _, v := range x.attrs {
		if !v.Allowed {
			return x.attrs, false
		}
	}
	return x.attrs, true
}

func ExtractResourceAttributes(data []byte, verb string, reg *hub.Registry, attrs map[authorization.ResourceAttributes]*ResourcePermission) error {
	reader := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 2048)
	for {
		var obj unstructured.Unstructured
		err := reader.Decode(&obj)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if obj.IsList() {
			err := obj.EachListItem(func(item runtime.Object) error {
				castItem := item.(*unstructured.Unstructured)

				gvr, err := reg.GVR(schema.FromAPIVersionAndKind(castItem.GetAPIVersion(), castItem.GetKind()))
				if err != nil {
					return err
				}

				ns := XorY(castItem.GetNamespace(), core.NamespaceDefault)
				castItem.SetNamespace(ns)

				attr := authorization.ResourceAttributes{
					Namespace: ns,
					Verb:      verb,
					Group:     gvr.Group,
					Version:   gvr.Version,
					Resource:  gvr.Resource,
					// Name:      castItem.GetName(), // TODO: needed for delete
				}
				info, found := attrs[attr]
				if !found {
					info = new(ResourcePermission)
					attrs[attr] = info
				}
				info.Items = append(info.Items, castItem)

				return nil
			})
			if err != nil {
				return err
			}
		} else {
			gvr, err := reg.GVR(schema.FromAPIVersionAndKind(obj.GetAPIVersion(), obj.GetKind()))
			if err != nil {
				return err
			}

			ns := XorY(obj.GetNamespace(), core.NamespaceDefault)
			obj.SetNamespace(ns)

			attr := authorization.ResourceAttributes{
				Namespace: ns,
				Verb:      verb,
				Group:     gvr.Group,
				Version:   gvr.Version,
				Resource:  gvr.Resource,
				// Name:      obj.GetName(), // TODO: needed for delete
			}
			info, found := attrs[attr]
			if !found {
				info = new(ResourcePermission)
				attrs[attr] = info
			}
			info.Items = append(info.Items, &obj)
		}
	}
	return nil
}

func ExtractResources(data []byte) ([]*unstructured.Unstructured, error) {
	var resources []*unstructured.Unstructured

	reader := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 2048)
	for {
		var obj unstructured.Unstructured
		err := reader.Decode(&obj)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if obj.IsList() {
			err := obj.EachListItem(func(item runtime.Object) error {
				castItem := item.(*unstructured.Unstructured)
				if castItem.GetNamespace() == "" {
					castItem.SetNamespace(core.NamespaceDefault)
				}
				resources = append(resources, castItem)
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if obj.GetNamespace() == "" {
				obj.SetNamespace(core.NamespaceDefault)
			}
			resources = append(resources, &obj)
		}
	}

	return resources, nil
}

// helm.sh/helm/v3/pkg/action/install.go
const notesFileSuffix = "NOTES.txt"

// renderResources renders the templates in a chart
func renderResources(ch *chart.Chart, caps *chartutil.Capabilities, values chartutil.Values) ([]*release.Hook, []releaseutil.Manifest, error) {
	hs := []*release.Hook{}
	b := bytes.NewBuffer(nil)

	if ch.Metadata.KubeVersion != "" {
		if !chartutil.IsCompatibleRange(ch.Metadata.KubeVersion, caps.KubeVersion.String()) {
			return hs, nil, errors.Errorf("chart requires kubeVersion: %s which is incompatible with Kubernetes %s", ch.Metadata.KubeVersion, caps.KubeVersion.String())
		}
	}

	files, err := engine.Render(ch, values)
	if err != nil {
		return hs, nil, err
	}

	for k := range files {
		if strings.HasSuffix(k, notesFileSuffix) {
			delete(files, k)
		}
	}

	// Sort hooks, manifests, and partials. Only hooks and manifests are returned,
	// as partials are not used after renderer.Render. Empty manifests are also
	// removed here.
	hs, manifests, err := releaseutil.SortManifests(files, caps.APIVersions, releaseutil.InstallOrder)
	if err != nil {
		// By catching parse errors here, we can prevent bogus releases from going
		// to Kubernetes.
		//
		// We return the files as a big blob of data to help the user debug parser
		// errors.
		for name, content := range files {
			if strings.TrimSpace(content) == "" {
				continue
			}
			fmt.Fprintf(b, "---\n# Source: %s\n%s\n", name, content)
		}
		return hs, manifests, err
	}

	return hs, manifests, nil
}

func IsEvent(events []release.HookEvent, x release.HookEvent) bool {
	for _, event := range events {
		if event == x {
			return true
		}
	}
	return false
}

// isChartInstallable validates if a chart can be installed
//
// Application chart type is only installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// url -> name
var repos = map[string]string{}
var repoLock sync.Mutex

func init() {
	go func() {
		err := wait.PollImmediateInfinite(24*time.Hour, func() (done bool, err error) {
			resp, err := http.Get("https://raw.githubusercontent.com/helm/hub/master/repos.yaml")
			if err == nil {
				defer resp.Body.Close()
				if data, err := ioutil.ReadAll(resp.Body); err == nil {
					var hub v1alpha1.Hub
					err = yamllib.Unmarshal(data, &hub)
					if err == nil {
						repoLock.Lock()
						for _, repo := range hub.Repositories {
							repos[strings.TrimSuffix(repo.URL, "/")] = repo.Name
						}
						repoLock.Unlock()
					}
				}
			}
			return false, nil // never exit
		})
		if err != nil {
			panic(err)
		}
	}()
}

func getCachedRepoName(chartURL string) (string, bool) {
	repoLock.Lock()
	defer repoLock.Unlock()
	name, ok := repos[strings.TrimSuffix(chartURL, "/")]
	return name, ok
}

func RepoName(chartURL string) (string, error) {
	name, ok := getCachedRepoName(chartURL)
	if ok {
		return name, nil
	}

	u, err := url.Parse(chartURL)
	if err != nil {
		return "", err
	}

	hostname := u.Hostname()
	ip := net.ParseIP(hostname)
	if ip == nil {
		name = repoNameFromDomain(hostname)
		repoLock.Lock()
		repos[strings.TrimSuffix(chartURL, "/")] = name
		repoLock.Unlock()
		return name, nil
	} else if ipv4 := ip.To4(); ipv4 != nil {
		return strings.ReplaceAll(ipv4.String(), ".", "-"), nil
	} else if ipv6 := ip.To16(); ipv6 != nil {
		return strings.ReplaceAll(ipv6.String(), ":", "-"), nil
	}
	return "", fmt.Errorf("failed to generate repo name for url:%s", chartURL)
}

func repoNameFromDomain(domain string) string {
	// TODO: Use https://raw.githubusercontent.com/helm/hub/master/repos.yaml
	if strings.HasSuffix(domain, ".storage.googleapis.com") {
		return strings.TrimSuffix(domain, ".storage.googleapis.com")
	}

	publicSuffix, icann := publicsuffix.PublicSuffix(domain)
	if icann {
		domain = strings.TrimSuffix(domain, "."+publicSuffix)
	}
	if strings.HasPrefix(domain, "charts.") {
		domain = strings.TrimPrefix(domain, "charts.")
	}

	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts)/2; i++ {
		j := len(parts) - i - 1
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "-")
}
