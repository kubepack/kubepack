package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"unicode"

	gomime "github.com/cubewise-code/go-mime"
	"github.com/gabriel-vasile/mimetype"
	"helm.sh/helm/v3/pkg/chart"
	ylib "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type DataFormat string

const (
	JsonFormat DataFormat = "json"
	YAMLFormat DataFormat = "yaml"
)

func Marshal(v interface{}, format DataFormat) ([]byte, error) {
	if format == JsonFormat {
		return json.Marshal(v)
	} else if format == YAMLFormat {
		return yaml.Marshal(v)
	}
	return nil, fmt.Errorf("unknonw format: %v", format)
}

func ConvertFormat(f *chart.File, format DataFormat) ([]byte, string, error) {
	if format == JsonFormat {
		out, err := ylib.ToJSON(f.Data)
		if err != nil {
			return nil, "", err
		}
		return out, "application/json", nil
	} else if format == YAMLFormat {
		if hasJSONPrefix(f.Data) {
			out, err := yaml.JSONToYAML(f.Data)
			if err != nil {
				return nil, "", err
			}
			return out, "text/yaml", nil
		}
	}
	ext := filepath.Ext(f.Name)
	if ct := gomime.TypeByExtension(ext); ct != "" {
		return f.Data, ct, nil
	}

	ct := mimetype.Detect(f.Data)
	return f.Data, ct.String(), nil
}

var jsonPrefix = []byte("{")

// hasJSONPrefix returns true if the provided buffer appears to start with
// a JSON open brace.
func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// Return true if the first non-whitespace bytes in buf is
// prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}

func ConvertChartTemplates(tpls []ChartTemplate, format DataFormat) ([]ChartTemplateOutput, error) {
	var out []ChartTemplateOutput

	for _, tpl := range tpls {
		entry := ChartTemplateOutput{
			ChartRef:    tpl.ChartRef,
			Version:     tpl.Version,
			ReleaseName: tpl.ReleaseName,
			Namespace:   tpl.Namespace,
			Manifest:    tpl.Manifest,
		}

		for _, crd := range tpl.CRDs {
			data, err := Marshal(crd, format)
			if err != nil {
				return nil, err
			}
			entry.CRDs = append(entry.CRDs, BucketFileOutput{
				URL:      crd.URL,
				Key:      crd.Key,
				Filename: crd.Filename,
				Data:     string(data),
			})
		}
		for _, r := range tpl.Resources {
			data, err := Marshal(r, format)
			if err != nil {
				return nil, err
			}
			entry.Resources = append(entry.Resources, string(data))
		}
		out = append(out, entry)
	}

	return out, nil
}
