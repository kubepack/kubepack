package action

import (
	"fmt"
	"log"
	"strings"

	"github.com/gobuffalo/flect"
	"helm.sh/helm/v3/pkg/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"kmodules.xyz/resource-metadata/hub"
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

func RefillMetadata(reg *hub.Registry, ref, actual map[string]interface{}, gvr metav1.GroupVersionResource, rls types.NamespacedName) error {
	actual["metadata"] = map[string]interface{}{
		"resource": map[string]interface{}{
			"group":    gvr.Group,
			"version":  gvr.Version,
			"resource": gvr.Resource,
		},
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

	for key, o := range actualResources {
		// apiVersion
		// kind
		// metadata:
		//	name:
		//  namespace:
		//  labels:

		refObj, ok := refResources[key].(map[string]interface{})
		if !ok {
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
		if gvr, err := reg.GVR(gvk); err == nil {
			if rd, err := reg.LoadByGVR(gvr); err == nil {
				if rd.Spec.UI != nil {
					for _, fields := range rd.Spec.UI.InstanceLabelPaths {
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
