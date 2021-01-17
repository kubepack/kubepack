package lib

import (
	"kubepack.dev/lib-app/api"

	meta_util "kmodules.xyz/client-go/meta"
)

func ConvertChartTemplates(tpls []api.ChartTemplate, format meta_util.DataFormat) ([]api.ChartTemplateOutput, error) {
	var out []api.ChartTemplateOutput

	for _, tpl := range tpls {
		entry := api.ChartTemplateOutput{
			ChartRef:    tpl.ChartRef,
			Version:     tpl.Version,
			ReleaseName: tpl.ReleaseName,
			Namespace:   tpl.Namespace,
			Manifest:    tpl.Manifest,
		}

		for _, crd := range tpl.CRDs {
			data, err := meta_util.Marshal(crd, format)
			if err != nil {
				return nil, err
			}
			entry.CRDs = append(entry.CRDs, api.BucketFileOutput{
				URL:      crd.URL,
				Key:      crd.Key,
				Filename: crd.Filename,
				Data:     string(data),
			})
		}
		for _, r := range tpl.Resources {
			data, err := meta_util.Marshal(r, format)
			if err != nil {
				return nil, err
			}
			entry.Resources = append(entry.Resources, string(data))
		}
		out = append(out, entry)
	}

	return out, nil
}
