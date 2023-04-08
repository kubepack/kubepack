package values

import (
	"context"

	"helm.sh/helm/v3/pkg/chart"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	chartsapi "x-helm.dev/apimachinery/apis/charts/v1alpha1"
)

func MergePresetValues(kc client.Client, chrt *chart.Chart, ref chartsapi.ChartPresetRef) (map[string]interface{}, error) {
	var valOpts Options
	if ref.PresetName != "" {
		ps, err := ref.ClusterChartPreset()
		if err != nil {
			return nil, err
		}
		valOpts, err = LoadClusterChartPresetValues(kc, ps, ref.Namespace)
		if err != nil {
			return nil, err
		}
	}
	return valOpts.MergeValues(chrt)
}

func LoadClusterChartPresetValues(kc client.Client, in chartsapi.Preset, ns string) (Options, error) {
	sel, err := metav1.LabelSelectorAsSelector(in.GetSpec().Selector)
	if err != nil {
		return Options{}, err
	}

	var opts Options
	err = mergeClusterChartPresetValues(kc, in, ns, sel, &opts)
	if err != nil {
		return Options{}, err
	}
	return opts, nil
}

func mergeClusterChartPresetValues(kc client.Client, in chartsapi.Preset, ns string, sel labels.Selector, opts *Options) error {
	for _, ps := range in.GetSpec().UsePresets {
		if ps.Kind == chartsapi.ResourceKindClusterChartPreset {
			obj, err := getPreset(kc, ps, ns)
			if err != nil {
				return err
			}
			if sel.Matches(labels.Set(obj.GetLabels())) {
				if err := mergeClusterChartPresetValues(kc, obj, ns, sel, opts); err != nil {
					return err
				}
			}
		}
	}
	if in.GetSpec().Values != nil && in.GetSpec().Values.Raw != nil {
		opts.ValueBytes = append(opts.ValueBytes, in.GetSpec().Values.Raw)
	}
	return nil
}

func getPreset(kc client.Client, in core.TypedLocalObjectReference, ns string) (chartsapi.Preset, error) {
	// Usually namespace is set nby user for Options chart values
	if ns != "" {
		var cp chartsapi.ChartPreset
		err := kc.Get(context.TODO(), client.ObjectKey{Namespace: ns, Name: in.Name}, &cp)
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		} else if err == nil {
			return &cp, nil
		}
	}
	var ccp chartsapi.ClusterChartPreset
	err := kc.Get(context.TODO(), client.ObjectKey{Name: in.Name}, &ccp)
	return &ccp, err
}
