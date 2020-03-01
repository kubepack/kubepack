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
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/client/clientset/versioned"
	chart2 "kubepack.dev/lib-helm/chart"
	"kubepack.dev/lib-helm/repo"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
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
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"kmodules.xyz/client-go/apiextensions/v1beta1"
	wait2 "kmodules.xyz/client-go/tools/wait"
	"kmodules.xyz/resource-metadata/hub"
	yamllib "sigs.k8s.io/yaml"
)

type DoFn func() error

type NamespacePrinter struct {
	Namespace string
	W         io.Writer
}

func (x *NamespacePrinter) Do() error {
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
	WaitFors  []v1alpha1.WaitFlags
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
	Namespace string
	WaitFors  []v1alpha1.WaitFlags

	ClientGetter genericclioptions.RESTClientGetter
}

func (x *WaitForChecker) Do() error {
	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	for _, flags := range x.WaitFors {
		builder := resource.NewBuilder(x.ClientGetter).
			NamespaceParam(x.Namespace).DefaultNamespace().
			// AllNamespaces(true).
			Unstructured().
			Latest().
			ContinueOnError().
			Flatten()
		if flags.All {
			builder.ResourceTypeOrNameArgs(true, flags.Resource.Group)
		} else {
			builder.ResourceTypeOrNameArgs(false, flags.Resource.Group+"/"+flags.Resource.Name)
		}
		if flags.Labels != nil {
			selector, err := v1.LabelSelectorAsSelector(flags.Labels)
			if err != nil {
				return err
			}
			builder.LabelSelectorParam(selector.String())
		}

		clientConfig, err := x.ClientGetter.ToRESTConfig()
		if err != nil {
			return err
		}
		dynamicClient, err := dynamic.NewForConfig(clientConfig)
		if err != nil {
			return err
		}
		conditionFn, err := wait2.ConditionFuncFor(flags.ForCondition, streams.ErrOut)
		if err != nil {
			return err
		}

		effectiveTimeout := flags.Timeout.Duration
		if effectiveTimeout < 0 {
			effectiveTimeout = 168 * time.Hour
		}

		o := &wait2.WaitOptions{
			ResourceFinder: &ResourceFindBuilderWrapper{builder},
			DynamicClient:  dynamicClient,
			Timeout:        effectiveTimeout,

			Printer:     printers.NewDiscardingPrinter(),
			ConditionFn: conditionFn,
			IOStreams:   streams,
		}

		err = o.WaitUntilAvailable(flags.ForCondition)
		if err != nil {
			return err
		}
		err = o.RunWait()
		if err != nil {
			return err
		}
	}
	return nil
}

// ResourceFindBuilderWrapper wraps a builder in an interface
type ResourceFindBuilderWrapper struct {
	builder *resource.Builder
}

// Do finds you resources to check
func (b *ResourceFindBuilderWrapper) Do() resource.Visitor {
	return b.builder.Do()
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
		// Work around for bug: https://github.com/kubernetes/kubernetes/issues/83242
		_, err := fmt.Fprintf(x.W, "until kubectl get crds %s.%s -o=jsonpath='{.items[0].metadata.name}' &> /dev/null; do sleep 1; done\n", crd.Name, crd.Group)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(x.W, "kubectl wait --for=condition=Established crds/%s.%s --timeout=5m\n", crd.Name, crd.Group)
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
	Registry    *repo.Registry
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	ValuesFile  string
	ValuesPatch *runtime.RawExtension

	W io.Writer
}

const indent = "  "

func (x *Helm3CommandPrinter) Do() error {
	chrt, err := x.Registry.GetChart(x.ChartRef.URL, x.ChartRef.Name, x.Version)
	if err != nil {
		return err
	}

	reponame, err := repo.DefaultNamer.Name(x.ChartRef.URL)
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
		vals := chrt.Values

		if x.ValuesFile != "" {
			for _, f := range chrt.Files {
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
		var modified map[string]interface{}
		err = json.Unmarshal(modifiedValues, &modified)
		if err != nil {
			return err
		}
		setValues := chart2.GetChangedValues(chrt.Values, modified)
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
	Registry    *repo.Registry
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	ValuesFile  string
	ValuesPatch *runtime.RawExtension

	W io.Writer
}

func (x *Helm2CommandPrinter) Do() error {
	chrt, err := x.Registry.GetChart(x.ChartRef.URL, x.ChartRef.Name, x.Version)
	if err != nil {
		return err
	}

	reponame, err := repo.DefaultNamer.Name(x.ChartRef.URL)
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
		vals := chrt.Values

		if x.ValuesFile != "" {
			for _, f := range chrt.Files {
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
		var modified map[string]interface{}
		err = json.Unmarshal(modifiedValues, &modified)
		if err != nil {
			return err
		}
		setValues := chart2.GetChangedValues(chrt.Values, modified)
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

type YAMLPrinter struct {
	Registry    *repo.Registry
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	KubeVersion string
	ValuesFile  string
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

	if chrt.Metadata.Deprecated {
		_, err = fmt.Fprintln(&buf, "# WARNING: This chart is deprecated")
		if err != nil {
			return err
		}

	}

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
			for _, f := range chrt.Files {
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

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	_ = log.Output(2, fmt.Sprintf(format, v...))
}

type ResourcePermission struct {
	Items   []*unstructured.Unstructured
	Allowed bool
}

type PermissionChecker struct {
	Registry    *repo.Registry
	ChartRef    v1alpha1.ChartRef
	Version     string
	ReleaseName string
	Namespace   string
	Verb        string

	Config           *rest.Config
	ClientGetter     genericclioptions.RESTClientGetter
	ResourceRegistry *hub.Registry

	attrs map[authorization.ResourceAttributes]*ResourcePermission
	m     sync.Mutex
}

func (x *PermissionChecker) Do() error {
	if x.attrs == nil {
		x.attrs = make(map[authorization.ResourceAttributes]*ResourcePermission)
	}

	chrt, err := x.Registry.GetChart(x.ChartRef.URL, x.ChartRef.Name, x.Version)
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

	validInstallableChart, err := isChartInstallable(chrt.Chart)
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

	if req := chrt.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chrt.Chart, req); err != nil {
			return err
		}
	}

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

	if err := chartutil.ProcessDependencies(chrt.Chart, vals); err != nil {
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
	valuesToRender, err := chartutil.ToRenderValues(chrt.Chart, vals, options, caps)
	if err != nil {
		return err
	}

	hooks, manifests, err := renderResources(chrt.Chart, caps, valuesToRender)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPreInstall) {
			err = ExtractResourceAttributes([]byte(hook.Manifest), x.Verb, x.ResourceRegistry, x.attrs)
			if err != nil {
				return err
			}
		}
	}

	for _, m := range manifests {
		err = ExtractResourceAttributes([]byte(m.Content), x.Verb, x.ResourceRegistry, x.attrs)
		if err != nil {
			return err
		}
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPostInstall) {
			err = ExtractResourceAttributes([]byte(hook.Manifest), x.Verb, x.ResourceRegistry, x.attrs)
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

type ApplicationCRDRegPrinter struct {
	W io.Writer
}

func (x *ApplicationCRDRegPrinter) Do() error {
	_, err := fmt.Fprintln(x.W, "kubectl apply -f https://github.com/kubepack/kubepack/raw/prototype/api/crds/kubepack.com_applications.yaml")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(x.W, "kubectl wait --for=condition=Established crds/applications.kubepack.com --timeout=5m")
	if err != nil {
		return err
	}
	return nil
}

type ApplicationCRDRegistrar struct {
	Config *rest.Config
}

func (x *ApplicationCRDRegistrar) Do() error {
	kc, err := kubernetes.NewForConfig(x.Config)
	if err != nil {
		return err
	}
	apiextClient, err := crd_cs.NewForConfig(x.Config)
	if err != nil {
		return err
	}
	return v1beta1.RegisterCRDs(kc.Discovery(), apiextClient, []*crdv1beta1.CustomResourceDefinition{
		v1alpha1.Application{}.CustomResourceDefinition(),
	})
}

type ApplicationUploader struct {
	App       *v1alpha1.Application
	UID       string
	BucketURL string
	PublicURL string
	W         io.Writer
}

func (x *ApplicationUploader) Do() error {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, x.BucketURL)
	if err != nil {
		return err
	}

	bucket = blob.PrefixedBucket(bucket, x.UID+"/apps/"+x.App.Namespace+"/")
	defer bucket.Close()

	data, err := yamllib.Marshal(x.App)
	if err != nil {
		return err
	}

	w, err := bucket.NewWriter(ctx, x.App.Name+".yaml", nil)
	if err != nil {
		return err
	}
	_, writeErr := fmt.Fprintln(w, string(data))
	// Always check the return value of Close when writing.
	closeErr := w.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}

	_, err = fmt.Fprintf(x.W, "kubectl apply -f %s/%s\n", x.PublicURL, path.Join(x.UID, "apps", x.App.Namespace, x.App.Name+".yaml"))
	if err != nil {
		return err
	}
	return nil
}

type ApplicationCreator struct {
	App    *v1alpha1.Application
	Client *versioned.Clientset
}

func (x *ApplicationCreator) Do() error {
	_, err := x.Client.KubepackV1alpha1().Applications(x.App.Namespace).Create(x.App)
	return err
}

type ApplicationGenerator struct {
	Registry *repo.Registry
	Chart    v1alpha1.ChartSelection
	chrt     *chart.Chart

	KubeVersion string

	components   map[metav1.GroupKind]struct{}
	commonLabels map[string]string
	init         bool
}

func (x *ApplicationGenerator) Do() error {
	if x.components == nil {
		x.components = make(map[metav1.GroupKind]struct{})
	}
	if x.commonLabels == nil {
		x.commonLabels = make(map[string]string)
	}

	chrt, err := x.Registry.GetChart(x.Chart.URL, x.Chart.Name, x.Chart.Version)
	x.chrt = chrt.Chart
	if err != nil {
		return err
	}

	cfg := new(action.Configuration)
	//err = cfg.Init(x.ClientGetter, x.Namespace, "memory", debug)
	//if err != nil {
	//	return err
	//}

	client := action.NewInstall(cfg)
	var extraAPIs []string

	client.DryRun = true
	client.ReleaseName = x.Chart.ReleaseName
	client.Namespace = x.Chart.Namespace
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.APIVersions = chartutil.VersionSet(extraAPIs)
	client.Version = x.Chart.Version

	validInstallableChart, err := isChartInstallable(x.chrt)
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

	if req := x.chrt.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(x.chrt, req); err != nil {
			return err
		}
	}

	vals := x.chrt.Values
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
	//if crds := chrt.CRDs(); len(crds) > 0 {
	//	attr := metav1.GroupKind{
	//		Group:    "apiextensions.k8s.io",
	//		Kind: "CustomResourceDefinition",
	//	}
	//	x.components[attr] = struct{}{}
	//}

	if err := chartutil.ProcessDependencies(x.chrt, vals); err != nil {
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
		Name:      x.Chart.ReleaseName,
		Namespace: x.Chart.Namespace,
		Revision:  1,
		IsInstall: true,
	}
	valuesToRender, err := chartutil.ToRenderValues(x.chrt, vals, options, caps)
	if err != nil {
		return err
	}

	hooks, manifests, err := renderResources(x.chrt, caps, valuesToRender)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPreInstall) {
			err = x.extractComponentAttributes([]byte(hook.Manifest))
			if err != nil {
				return err
			}
		}
	}

	for _, m := range manifests {
		err = x.extractComponentAttributes([]byte(m.Content))
		if err != nil {
			return err
		}
	}

	for _, hook := range hooks {
		if IsEvent(hook.Events, release.HookPostInstall) {
			err = x.extractComponentAttributes([]byte(hook.Manifest))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (x *ApplicationGenerator) Result() *v1alpha1.Application {
	desc := GetPackageDescriptor(x.chrt)

	b := &v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindApplication,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        x.Chart.ReleaseName,
			Namespace:   x.Chart.Namespace,
			Labels:      nil, // TODO: ?
			Annotations: nil, // TODO: ?
		},
		Spec: v1alpha1.ApplicationSpec{
			Description: v1alpha1.Descriptor{
				PackageDescriptor: v1alpha1.PackageDescriptor{
					Type:        x.chrt.Name(),
					Description: desc.Description,
					Icons:       desc.Icons,
					Maintainers: desc.Maintainers,
					Keywords:    desc.Keywords,
					Links:       desc.Links,
					Notes:       "",
				},
				Version: x.chrt.Metadata.AppVersion,
				Owners:  nil, // TODO: Add the user email who is installing this app
			},
			AddOwnerRef:   false,
			Info:          nil,
			AssemblyPhase: v1alpha1.Ready,
			Package: v1alpha1.ApplicationPackage{
				Bundle: x.Chart.Bundle,
				Chart: v1alpha1.ChartRepoRef{
					Name:    x.Chart.Name,
					URL:     x.Chart.URL,
					Version: x.Chart.Version,
				},
				Channel: v1alpha1.RegularChannel,
			},
		},
	}

	gks := make([]metav1.GroupKind, 0, len(x.components))
	for gk := range x.components {
		gks = append(gks, gk)
	}
	sort.Slice(gks, func(i, j int) bool {
		if gks[i].Group == gks[j].Group {
			return gks[i].Kind < gks[j].Kind
		}
		return gks[i].Group < gks[j].Group
	})
	b.Spec.ComponentGroupKinds = gks

	if len(x.commonLabels) > 0 {
		b.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: x.commonLabels,
		}
	}
	return b
}

func (x *ApplicationGenerator) extractComponentAttributes(data []byte) error {
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

				gv, err := schema.ParseGroupVersion(castItem.GetAPIVersion())
				if err != nil {
					return err
				}
				x.components[metav1.GroupKind{Group: gv.Group, Kind: castItem.GetKind()}] = struct{}{}

				if !x.init {
					x.commonLabels = castItem.GetLabels()
					x.init = true
				} else {
					for k, v := range castItem.GetLabels() {
						if existing, found := x.commonLabels[k]; found && existing != v {
							delete(x.commonLabels, k)
						}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())
			if err != nil {
				return err
			}
			x.components[metav1.GroupKind{Group: gv.Group, Kind: obj.GetKind()}] = struct{}{}

			if !x.init {
				x.commonLabels = obj.GetLabels()
				x.init = true
			} else {
				for k, v := range obj.GetLabels() {
					if existing, found := x.commonLabels[k]; found && existing != v {
						delete(x.commonLabels, k)
					}
				}
			}
		}
	}
	return nil
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
