package values

import (
	"encoding/json"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

/*
- If ReplaceValues != nil, then just use that values as=is
- else
-   if ValuesPatch != nil, then apply patch to the ValuesFile and use the generated values
-   else coalesce the StringValues and Values into ValuesFile and use those as the final values.

See also: https://github.com/helm/helm/blob/v3.5.4/pkg/cli/values/options.go#L39-L86
*/
type Options struct {
	ReplaceValues map[string]any        `json:"replaceValues"`
	ValuesFile    string                `json:"valuesFile"`
	ValuesPatch   *runtime.RawExtension `json:"valuesPatch"`
	ValueBytes    [][]byte              `json:"valueBytes"`
	StringValues  []string              `json:"stringValues"`
	Values        []string              `json:"values"`
	KVPairs       []KV                  `json:"kv_pairs"`
}

type KV struct {
	K string
	V any
}

// MergeValues merges values from files specified via -f/--values and directly
// via --set, --set-string, or --set-file, marshaling them to YAML
func (opts *Options) MergeValues(chrt *chart.Chart) (map[string]any, error) {
	// Note that len(opts.ReplaceValues) == 0 will be considered a valid replacement
	if opts.ReplaceValues != nil {
		return opts.ReplaceValues, nil
	}

	if opts.ValuesFile == "" {
		opts.ValuesFile = chartutil.ValuesfileName
	}

	var baseFile *chart.File
	for _, f := range chrt.Raw {
		if f.Name == opts.ValuesFile {
			baseFile = f
			break
		}
	}
	if baseFile == nil {
		return nil, fmt.Errorf("can't find values file %s", opts.ValuesFile)
	}

	if opts.ValuesPatch != nil {
		patchData, err := json.Marshal(opts.ValuesPatch)
		if err != nil {
			return nil, err
		}
		patch, err := jsonpatch.DecodePatch(patchData)
		if err != nil {
			return nil, err
		}

		baseBytes, err := yaml.YAMLToJSON(baseFile.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to convert values file %s to json, reason %v", opts.ValuesFile, err)
		}
		valuesBytes, err := patch.Apply(baseBytes)
		if err != nil {
			return nil, err
		}

		vals := map[string]any{}
		err = json.Unmarshal(valuesBytes, &vals)
		if err != nil {
			return nil, err
		}

		return vals, nil
	}

	// Use StringValues, Values flags
	base := map[string]any{}
	if err := yaml.Unmarshal(baseFile.Data, &base); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", opts.ValuesFile)
	}

	// User specified a values files via -f/--values
	for _, bytes := range opts.ValueBytes {
		currentMap := map[string]any{}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s", bytes)
		}
		// Merge with the previous map
		base = MergeMaps(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range opts.Values {
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, errors.Wrap(err, "failed parsing --set data")
		}
	}

	// User specified a value via --set-string
	for _, value := range opts.StringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return nil, errors.Wrap(err, "failed parsing --set-string data")
		}
	}

	for _, kv := range opts.KVPairs {
		err := unstructured.SetNestedField(base, kv.V, strings.Split(kv.K, ".")...)
		if err != nil {
			return nil, err
		}
	}

	return base, nil
}

func MergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
