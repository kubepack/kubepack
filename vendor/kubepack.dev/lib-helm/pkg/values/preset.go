package values

import (
	"context"

	"helm.sh/helm/v3/pkg/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	chartsapi "x-helm.dev/apimachinery/apis/charts/v1alpha1"
)

func MergePresetValues(kc client.Client, chrt *chart.Chart, ref chartsapi.ChartPresetFlatRef) (map[string]interface{}, error) {
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
	var opts Options
	err := mergeClusterChartPresetValues(kc, in, ns, &opts)
	if err != nil {
		return Options{}, err
	}
	return opts, nil
}

func mergeClusterChartPresetValues(kc client.Client, in chartsapi.Preset, ns string, opts *Options) error {
	for _, ps := range in.GetSpec().UsePresets {
		var presets []chartsapi.Preset
		if ps.Name != "" {
			obj, err := getPreset(kc, ps, ns)
			if err != nil {
				return err
			}
			presets = append(presets, obj)
		} else if ps.Selector != nil {
			sel, err := metav1.LabelSelectorAsSelector(ps.Selector)
			if err != nil {
				return err
			}

			var list metav1.PartialObjectMetadataList
			list.SetGroupVersionKind(chartsapi.GroupVersion.WithKind(chartsapi.ResourceKindClusterChartPreset))
			err = kc.List(context.TODO(), &list, client.MatchingLabelsSelector{
				Selector: sel,
			})
			if err != nil {
				return err
			}
			for _, md := range list.Items {
				ref := chartsapi.TypedLocalObjectReference{
					APIGroup: &chartsapi.GroupVersion.Group,
					Kind:     chartsapi.ResourceKindClusterChartPreset,
					Name:     md.Name,
				}
				obj, err := getPreset(kc, ref, ns)
				if err != nil {
					return err
				}
				presets = append(presets, obj)
			}
		}

		for _, obj := range presets {
			if err := mergeClusterChartPresetValues(kc, obj, ns, opts); err != nil {
				return err
			}
		}
	}
	if in.GetSpec().Values != nil && in.GetSpec().Values.Raw != nil {
		opts.ValueBytes = append(opts.ValueBytes, in.GetSpec().Values.Raw)
	}
	return nil
}

func getPreset(kc client.Client, in chartsapi.TypedLocalObjectReference, ns string) (chartsapi.Preset, error) {
	// Usually namespace is set by user for Options chart values
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
