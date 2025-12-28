package values

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kmapi "kmodules.xyz/client-go/api/v1"
	clustermeta "kmodules.xyz/client-go/cluster"
	uiapi "kmodules.xyz/resource-metadata/apis/ui/v1alpha1"
	"kmodules.xyz/resource-metadata/hub/resourceeditors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	chartsapi "x-helm.dev/apimachinery/apis/charts/v1alpha1"
)

func LoadPresetValues(kc client.Client, ref chartsapi.ChartPresetFlatRef) ([]chartsapi.ChartPresetValues, error) {
	if ref.Variant == "" {
		// required for editor charts
		return nil, nil
	}

	rid := &kmapi.ResourceID{
		Group: ref.Group,
		Name:  ref.Resource,
		Kind:  ref.Kind,
	}
	rid, err := kmapi.ExtractResourceID(kc.RESTMapper(), *rid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to detect resource ID for %#v", *rid)
	}
	ed, err := resourceeditors.LoadByGVR(kc, rid.GroupVersionResource())
	if err != nil {
		return nil, err
	}

	var variant *uiapi.VariantRef
	for i := range ed.Spec.Variants {
		if ed.Spec.Variants[i].Name == ref.Variant {
			variant = &ed.Spec.Variants[i]
			break
		}
	}
	if variant == nil {
		return nil, errors.Errorf("No variant with name %s found for %+v", ref.Variant, *rid)
	}

	if variant.Selector == nil {
		return nil, nil
	}
	sel, err := metav1.LabelSelectorAsSelector(variant.Selector)
	if err != nil {
		return nil, err
	}

	if ref.Namespace == "" {
		return bundleClusterChartPresets(kc, sel, nil)
	}

	values, err := bundleChartPresets(kc, ref.Namespace, sel, nil)
	if err != nil {
		return nil, err
	}
	knownPresets := map[string]bool{} // true => cp, false => ccp
	for _, v := range values {
		knownPresets[v.Source.Ref.Name] = true
	}

	if clustermeta.IsRancherManaged(kc.RESTMapper()) {
		projectId, found, err := clustermeta.GetProjectId(kc, ref.Namespace)
		if err != nil {
			return nil, err
		}
		if !found {
			// NS not in a project. So, just add the extra CCPs
			ccps, err := bundleClusterChartPresets(kc, sel, knownPresets)
			if err != nil {
				return nil, err
			}
			values = append(ccps, values...)
			return values, nil
		}

		nsList, err := clustermeta.ListProjectNamespaces(kc, projectId)
		if err != nil {
			return nil, err
		}
		projectPresets := map[string]bool{} // true => cp, false => ccp
		namespaces := clustermeta.Names(nsList)
		for i := len(namespaces) - 1; i >= 0; i-- {
			if namespaces[i] == ref.Namespace {
				continue
			}

			nsPresets, err := bundleChartPresets(kc, namespaces[i], sel, knownPresets)
			if err != nil {
				return nil, err
			}
			for _, v := range nsPresets {
				projectPresets[v.Source.Ref.Name] = true
			}
			values = append(nsPresets, values...)
		}
		// mark project presets as known
		for k, v := range projectPresets {
			knownPresets[k] = v
		}

		ccps, err := bundleClusterChartPresets(kc, sel, knownPresets)
		if err != nil {
			return nil, err
		}
		values = append(ccps, values...)
	} else {
		ccps, err := bundleClusterChartPresets(kc, sel, knownPresets)
		if err != nil {
			return nil, err
		}
		values = append(ccps, values...)
	}

	return values, nil
}

func bundleChartPresets(kc client.Client, ns string, sel labels.Selector, knownPresets map[string]bool) ([]chartsapi.ChartPresetValues, error) {
	var list chartsapi.ChartPresetList
	err := kc.List(context.TODO(), &list, client.InNamespace(ns), client.MatchingLabelsSelector{Selector: sel})
	if err != nil {
		return nil, err
	}
	cps := list.Items
	sort.Slice(cps, func(i, j int) bool {
		return cps[i].Name < cps[j].Name
	})

	values := make([]chartsapi.ChartPresetValues, 0, len(cps))
	for _, cp := range cps {
		if _, exists := knownPresets[cp.Name]; exists {
			continue
		}

		values = append(values, chartsapi.ChartPresetValues{
			Source: chartsapi.SourceLocator{
				Resource: kmapi.ResourceID{
					Group:   chartsapi.GroupVersion.Group,
					Version: chartsapi.GroupVersion.Version,
					Kind:    chartsapi.ResourceKindChartPreset,
				},
				Ref: kmapi.ObjectReference{
					Namespace: cp.Namespace,
					Name:      cp.Namespace,
				},
				UID:        cp.UID,
				Generation: cp.Generation,
			},
			Values: cp.Spec.Values,
		})
	}
	return values, err
}

func bundleClusterChartPresets(kc client.Client, sel labels.Selector, knownPresets map[string]bool) ([]chartsapi.ChartPresetValues, error) {
	var list chartsapi.ClusterChartPresetList
	err := kc.List(context.TODO(), &list, client.MatchingLabelsSelector{Selector: sel})
	if err != nil {
		return nil, err
	}

	ccps := list.Items
	sort.Slice(ccps, func(i, j int) bool {
		return ccps[i].Name < ccps[j].Name
	})

	values := make([]chartsapi.ChartPresetValues, 0, len(ccps))
	for _, ccp := range ccps {
		if _, exists := knownPresets[ccp.Name]; exists {
			continue
		}

		values = append(values, chartsapi.ChartPresetValues{
			Source: chartsapi.SourceLocator{
				Resource: kmapi.ResourceID{
					Group:   chartsapi.GroupVersion.Group,
					Version: chartsapi.GroupVersion.Version,
					Kind:    chartsapi.ResourceKindClusterChartPreset,
				},
				Ref: kmapi.ObjectReference{
					Namespace: ccp.Namespace,
					Name:      ccp.Namespace,
				},
				UID:        ccp.UID,
				Generation: ccp.Generation,
			},
			Values: ccp.Spec.Values,
		})
	}
	return values, nil
}

func MergePresetValues(kc client.Client, chrt *chart.Chart, ref chartsapi.ChartPresetFlatRef) (map[string]any, error) {
	presets, err := LoadPresetValues(kc, ref)
	if err != nil {
		return nil, err
	}

	var valOpts Options
	for _, preset := range presets {
		valOpts.ValueBytes = append(valOpts.ValueBytes, preset.Values.Raw)
	}
	return valOpts.MergeValues(chrt)
}
