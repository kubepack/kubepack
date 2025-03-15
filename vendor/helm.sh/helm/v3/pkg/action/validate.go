/*
Copyright The Helm Authors.

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

package action

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
)

var accessor = meta.NewAccessor()

const (
	appManagedByLabel              = "app.kubernetes.io/managed-by"
	appPartOfLabel                 = "app.kubernetes.io/part-of"
	appNameLabel                   = "app.kubernetes.io/name"
	appInstanceLabel               = "app.kubernetes.io/instance"
	editorLabel                    = "meta.x-helm.dev/editor"
	appManagedByHelm               = "Helm"
	helmReleaseNameAnnotation      = "meta.helm.sh/release-name"
	helmReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
)

// requireAdoption returns the subset of resources that already exist in the cluster.
func requireAdoption(resources kube.ResourceList) (kube.ResourceList, error) {
	var requireUpdate kube.ResourceList

	err := resources.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		helper := resource.NewHelper(info.Client, info.Mapping)
		_, err = helper.Get(info.Namespace, info.Name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return errors.Wrapf(err, "could not get information about the resource %s", resourceString(info))
		}

		requireUpdate.Append(info)
		return nil
	})

	return requireUpdate, err
}

func existingResourceConflict(resources kube.ResourceList, releaseName, releaseNamespace string, editorChart bool) (kube.ResourceList, error) {
	var requireUpdate kube.ResourceList

	err := resources.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		helper := resource.NewHelper(info.Client, info.Mapping)
		existing, err := helper.Get(info.Namespace, info.Name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return errors.Wrapf(err, "could not get information about the resource %s", resourceString(info))
		}

		if !editorChart {
			// Allow adoption of the resource if it is managed by Helm and is annotated with correct release name and namespace.
			if err := checkOwnership(existing, releaseName, releaseNamespace); err != nil {
				return fmt.Errorf("%s exists and cannot be imported into the current release: %s", resourceString(info), err)
			}
		}

		requireUpdate.Append(info)
		return nil
	})

	return requireUpdate, err
}

func checkOwnership(obj runtime.Object, releaseName, releaseNamespace string) error {
	lbls, err := accessor.Labels(obj)
	if err != nil {
		return err
	}
	annos, err := accessor.Annotations(obj)
	if err != nil {
		return err
	}

	var errs []error
	if err := requireValue(lbls, appManagedByLabel, appManagedByHelm); err != nil {
		errs = append(errs, fmt.Errorf("label validation error: %s", err))
	}
	if err := requireValue(annos, helmReleaseNameAnnotation, releaseName); err != nil {
		errs = append(errs, fmt.Errorf("annotation validation error: %s", err))
	}
	if err := requireValue(annos, helmReleaseNamespaceAnnotation, releaseNamespace); err != nil {
		errs = append(errs, fmt.Errorf("annotation validation error: %s", err))
	}

	// WARNING(tamal): checkOwnership does not need to check for prt-of, name, instance labels, as the ones above already cover it.

	if len(errs) > 0 {
		err := errors.New("invalid ownership metadata")
		for _, e := range errs {
			err = fmt.Errorf("%w; %s", err, e)
		}
		return err
	}

	return nil
}

func requireValue(meta map[string]string, k, v string) error {
	actual, ok := meta[k]
	if !ok {
		return fmt.Errorf("missing key %q: must be set to %q", k, v)
	}
	if actual != v {
		return fmt.Errorf("key %q must equal %q: current value is %q", k, v, actual)
	}
	return nil
}

// setMetadataVisitor adds release tracking metadata to all resources. If force is enabled, existing
// ownership metadata will be overwritten. Otherwise an error will be returned if any resource has an
// existing and conflicting value for the managed by label or Helm release/namespace annotations.
func setMetadataVisitor(releaseName, releaseNamespace string, extraLabels map[string]string, force bool) resource.VisitorFunc {
	return func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		if !force {
			if err := checkOwnership(info.Object, releaseName, releaseNamespace); err != nil {
				return fmt.Errorf("%s cannot be owned: %s", resourceString(info), err)
			}
		}

		if err := mergeLabels(info.Object, mergeStrStrMaps(
			map[string]string{
				appManagedByLabel: appManagedByHelm,
			},
			extraLabels,
		)); err != nil {
			return fmt.Errorf(
				"%s labels could not be updated: %s",
				resourceString(info), err,
			)
		}

		if err := mergeAnnotations(info.Object, map[string]string{
			helmReleaseNameAnnotation:      releaseName,
			helmReleaseNamespaceAnnotation: releaseNamespace,
		}); err != nil {
			return fmt.Errorf(
				"%s annotations could not be updated: %s",
				resourceString(info), err,
			)
		}

		return nil
	}
}

func resourceString(info *resource.Info) string {
	_, k := info.Mapping.GroupVersionKind.ToAPIVersionAndKind()
	return fmt.Sprintf(
		"%s %q in namespace %q",
		k, info.Name, info.Namespace,
	)
}

func mergeLabels(obj runtime.Object, labels map[string]string) error {
	current, err := accessor.Labels(obj)
	if err != nil {
		return err
	}
	return accessor.SetLabels(obj, mergeStrStrMaps(current, labels))
}

func mergeAnnotations(obj runtime.Object, annotations map[string]string) error {
	current, err := accessor.Annotations(obj)
	if err != nil {
		return err
	}
	return accessor.SetAnnotations(obj, mergeStrStrMaps(current, annotations))
}

// merge two maps, always taking the value on the right
func mergeStrStrMaps(current, desired map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range current {
		result[k] = v
	}
	for k, desiredVal := range desired {
		result[k] = desiredVal
	}
	return result
}

// special handling for appscode / kubepack specific case
func getAppLabels(rel *release.Release, cfg *Configuration) (map[string]string, error) {
	result := map[string]string{}
	// check storage driver name
	if cfg.Releases.Name() == "drivers.x-helm.dev/appreleases" {
		result[appInstanceLabel] = rel.Name

		if partOf, ok := rel.Chart.Metadata.Annotations[appPartOfLabel]; ok {
			result[appPartOfLabel] = partOf
		}
		if data, ok := rel.Chart.Metadata.Annotations[editorLabel]; ok && data != "" {
			var gvr metav1.GroupVersionResource
			if err := json.Unmarshal([]byte(data), &gvr); err != nil {
				return nil, errors.Wrapf(err, "failed to parse %s annotation value %s as GVR", editorLabel, data)
			}
			result[appNameLabel] = fmt.Sprintf("%s.%s", gvr.Resource, gvr.Group)
		}
	}
	return result, nil
}

func isEditorChart(ch *chart.Chart) bool {
	_, ok := ch.Metadata.Annotations[editorLabel]
	return ok
}
