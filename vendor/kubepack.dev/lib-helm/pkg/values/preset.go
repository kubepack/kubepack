package values

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"kmodules.xyz/client-go/tools/parser"
	chartsapi "kubepack.dev/preset/apis/charts/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MergePresetValues(kc client.Client, chrt *chart.Chart, ref chartsapi.ChartPresetRef) (map[string]interface{}, error) {
	vpsMap, err := LoadVendorPresets(chrt)
	if err != nil {
		return nil, err
	}

	var valOpts Options
	if ref.PresetName != "" {
		ps, err := ref.ClusterChartPreset()
		if err != nil {
			return nil, err
		}
		valOpts, err = LoadClusterChartPresetValues(kc, vpsMap, ps, ref.Namespace)
		if err != nil {
			return nil, err
		}
	}
	return valOpts.MergeValues(chrt)
}

func LoadClusterChartPresetValues(kc client.Client, vpsMap map[string]*chartsapi.VendorChartPreset, in chartsapi.Preset, ns string) (Options, error) {
	sel, err := metav1.LabelSelectorAsSelector(in.GetSpec().Selector)
	if err != nil {
		return Options{}, err
	}

	var opts Options
	err = mergeClusterChartPresetValues(kc, vpsMap, in, ns, sel, &opts)
	if err != nil {
		return Options{}, err
	}
	return opts, nil
}

func mergeClusterChartPresetValues(kc client.Client, vpsMap map[string]*chartsapi.VendorChartPreset, in chartsapi.Preset, ns string, sel labels.Selector, opts *Options) error {
	for _, ps := range in.GetSpec().UsePresets {
		if ps.Kind == chartsapi.ResourceKindVendorChartPreset {
			obj, ok := vpsMap[ps.Name]
			if !ok {
				return fmt.Errorf("missing VendorChartPreset %s", ps.Name)
			}
			if sel.Matches(labels.Set(obj.Labels)) {
				if err := mergeVendorChartPresetValues(vpsMap, sel, obj, opts); err != nil {
					return err
				}
			}
		} else if ps.Kind == chartsapi.ResourceKindClusterChartPreset {
			obj, err := getPreset(kc, in, ns)
			if err != nil {
				return err
			}
			if sel.Matches(labels.Set(obj.GetLabels())) {
				if err := mergeClusterChartPresetValues(kc, vpsMap, obj, ns, sel, opts); err != nil {
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

func getPreset(kc client.Client, in chartsapi.Preset, ns string) (chartsapi.Preset, error) {
	var cp chartsapi.ChartPreset
	err := kc.Get(context.TODO(), client.ObjectKey{Namespace: ns, Name: in.GetName()}, &cp)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	} else if err == nil {
		return &cp, nil
	}

	var ccp chartsapi.ClusterChartPreset
	err = kc.Get(context.TODO(), client.ObjectKey{Name: in.GetName()}, &ccp)
	return &ccp, err
}

func mergeVendorChartPresetValues(vpsMap map[string]*chartsapi.VendorChartPreset, sel labels.Selector, in *chartsapi.VendorChartPreset, opts *Options) error {
	for _, ps := range in.Spec.UsePresets {
		obj, ok := vpsMap[ps.Name]
		if !ok {
			return fmt.Errorf("missing VendorChartPreset %s", ps.Name)
		}
		if !sel.Matches(labels.Set(obj.Labels)) {
			continue
		}
		if err := mergeVendorChartPresetValues(vpsMap, sel, obj, opts); err != nil {
			return err
		}
	}
	if in.Spec.Values != nil && in.Spec.Values.Raw != nil {
		opts.ValueBytes = append(opts.ValueBytes, in.Spec.Values.Raw)
	}
	return nil
}

func LoadVendorPresets(chrt *chart.Chart) (map[string]*chartsapi.VendorChartPreset, error) {
	vpsMap := map[string]*chartsapi.VendorChartPreset{}
	for _, f := range chrt.Raw {
		if !strings.HasPrefix(f.Name, "presets/") {
			continue
		}
		if err := parser.ProcessResources(f.Data, func(ri parser.ResourceInfo) error {
			if ri.Object.GroupVersionKind() != chartsapi.GroupVersion.WithKind(chartsapi.ResourceKindVendorChartPreset) {
				return nil
			}

			var obj chartsapi.VendorChartPreset
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(ri.Object.UnstructuredContent(), &obj); err != nil {
				return errors.Wrapf(err, "failed to convert from unstructured obj %q in file %s", ri.Object.GetName(), ri.Filename)
			}
			vpsMap[obj.Name] = &obj

			return nil
		}); err != nil {
			return nil, err
		}

	}
	return vpsMap, nil
}
