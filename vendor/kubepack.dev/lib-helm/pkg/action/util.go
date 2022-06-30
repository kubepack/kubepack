package action

import (
	"fmt"
	"log"
	"strings"

	"github.com/gobuffalo/flect"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"kmodules.xyz/client-go/discovery"
	uiapi "kmodules.xyz/resource-metadata/apis/ui/v1alpha1"
	"kmodules.xyz/resource-metadata/hub/resourceeditors"
	chartsapi "kubepack.dev/preset/apis/charts/v1alpha1"
	storeapi "kubepack.dev/preset/apis/store/v1alpha1"
	appapi "sigs.k8s.io/application/api/app/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
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
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := chartsapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := storeapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := appapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := uiapi.AddToScheme(scheme); err != nil {
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
		return fmt.Errorf(".metadata.resource not found in ref values")
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
		name := rls.Name
		idx := strings.IndexRune(key, '_')
		if idx != -1 {
			name += "-" + flect.Dasherize(key[idx+1:])
		}
		err := unstructured.SetNestedField(obj, name, "metadata", "name")
		if err != nil {
			return err
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
