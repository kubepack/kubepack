package action

import (
	"fmt"
	"log"
	"sort"
	"strings"

	fluxhelm "github.com/fluxcd/helm-controller/api/v2"
	fluxsrc "github.com/fluxcd/source-controller/api/v1"
	"github.com/gobuffalo/flect"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	apiextensionsapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	apiregistrationapi "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"kmodules.xyz/client-go/discovery"
	uiapi "kmodules.xyz/resource-metadata/apis/ui/v1alpha1"
	"kmodules.xyz/resource-metadata/hub"
	"kmodules.xyz/resource-metadata/hub/resourceeditors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	chartsapi "x-helm.dev/apimachinery/apis/charts/v1alpha1"
	driversapi "x-helm.dev/apimachinery/apis/drivers/v1alpha1"
	productsapi "x-helm.dev/apimachinery/apis/products/v1alpha1"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	_ = log.Output(2, fmt.Sprintf(format, v...))
}

func setAnnotations(chrt *chart.Chart, k, v string) {
	if chrt.Metadata.Annotations == nil {
		chrt.Metadata.Annotations = map[string]string{}
	}
	if v != "" {
		chrt.Metadata.Annotations[k] = v
	} else {
		delete(chrt.Metadata.Annotations, k)
	}
}

func NewUncachedClient(getter action.RESTClientGetter) (client.Client, error) {
	cfg, err := getter.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	return NewUncachedClientForConfig(cfg)
}

func NewUncachedClientForConfig(cfg *rest.Config) (client.Client, error) {
	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := apiextensionsapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := apiregistrationapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// x-helm.dev
	if err := chartsapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := driversapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := productsapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := releasesapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// resource-metadata
	if err := uiapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// FluxCD
	if err := fluxsrc.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := fluxhelm.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}

func RefillMetadata(kc client.Client, ref, actual map[string]interface{}, gvr metav1.GroupVersionResource, rls types.NamespacedName) error {
	// WARNING: Don't use kc.RESTMapper().KindFor to find Kind because the CRD may be yet exist in the cluster
	rsMeta, ok, err := unstructured.NestedMap(ref, "metadata", "resource")
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf(".metadata.resource not found in chart values")
	}

	actual["metadata"] = map[string]interface{}{
		"resource": rsMeta,
		"release": map[string]interface{}{
			"name":      rls.Name,
			"namespace": rls.Namespace,
		},
	}

	refResources, ok := ref["resources"].(map[string]interface{})
	if !ok {
		return nil
	}
	actualResources, ok := actual["resources"].(map[string]interface{})
	if !ok {
		return nil
	}

	//rlsName, _, err := unstructured.NestedString(actual, "metadata", "release", "name")
	//if err != nil {
	//	return err
	//}
	//rlsNamespace, _, err := unstructured.NestedString(actual, "metadata", "release", "namespace")
	//if err != nil {
	//	return err
	//}
	mapper := discovery.NewResourceMapper(kc.RESTMapper())

	_, usesForm := ref["form"]

	for key, o := range actualResources {
		// apiVersion
		// kind
		// metadata:
		//	name:
		//  namespace:
		//  labels:

		refObj, ok := refResources[key].(map[string]interface{})
		if !ok {
			if usesForm {
				continue // in case of form, we will see form objects which are not present in the ref resources
			}
			return fmt.Errorf("missing key %s in reference chart values", key)
		}
		obj := o.(map[string]interface{})
		obj["apiVersion"] = refObj["apiVersion"]
		obj["kind"] = refObj["kind"]

		// name
		if !hub.IsFeaturesetGR(schema.GroupResource{Group: gvr.Group, Resource: gvr.Resource}) {
			name := rls.Name
			idx := strings.IndexRune(key, '_')
			if idx != -1 {
				name += "-" + flect.Dasherize(key[idx+1:])
			}
			err := unstructured.SetNestedField(obj, name, "metadata", "name")
			if err != nil {
				return err
			}
		}

		// namespace
		// TODO: add namespace if needed
		err = unstructured.SetNestedField(obj, rls.Namespace, "metadata", "namespace")
		if err != nil {
			return err
		}

		// get select labels from app and set to obj labels
		err = updateLabels(rls.Name, obj, "metadata", "labels")
		if err != nil {
			return err
		}

		gvk := schema.FromAPIVersionAndKind(refObj["apiVersion"].(string), refObj["kind"].(string))
		if gvr, err := mapper.GVR(gvk); err == nil {
			if ed, ok := resourceeditors.LoadByGVR(kc, gvr); ok {
				if ed.Spec.UI != nil {
					for _, fields := range ed.Spec.UI.InstanceLabelPaths {
						fields := strings.Trim(fields, ".")
						err = updateLabels(rls.Name, obj, strings.Split(fields, ".")...)
						if err != nil {
							return err
						}
					}
				}
			}
		}

		actualResources[key] = obj
	}
	return nil
}

func updateLabels(rlsName string, obj map[string]interface{}, fields ...string) error {
	labels, ok, err := unstructured.NestedStringMap(obj, fields...)
	if err != nil {
		return err
	}
	if !ok {
		labels = map[string]string{}
	}
	key := "app.kubernetes.io/instance"
	if _, ok := labels[key]; ok {
		labels[key] = rlsName
	}
	return unstructured.SetNestedStringMap(obj, labels, fields...)
}

func ExtractResourceKeys(vals map[string]interface{}) []string {
	if resources, ok, err := unstructured.NestedMap(vals, "resources"); err == nil && ok {
		resourceKeys := make([]string, 0, len(resources))
		for k := range resources {
			resourceKeys = append(resourceKeys, k)
		}
		sort.Strings(resourceKeys)
		return resourceKeys
	}
	return nil
}
