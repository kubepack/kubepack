package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"kubepack.dev/chart-doc-gen/api"
	"kubepack.dev/kubepack/pkg/lib"

	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/flect"
	"github.com/spf13/cobra"
	y3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	"kmodules.xyz/resource-metadata/hub"
	"sigs.k8s.io/yaml"
)

var (
	sampleDir   = ""
	chartDir    = ""
	chartName   = ""
	chartSchema = crdv1.JSONSchemaProps{
		Type:       "object",
		Properties: map[string]crdv1.JSONSchemaProps{},
	}
	modelValues  = map[string]ObjectContainer{}
	registry     = hub.NewRegistryOfKnownResources()
	resourceKeys = sets.NewString()
)

type ObjectModel struct {
	Key    string                     `json:"key"`
	Object *unstructured.Unstructured `json:"object"`
}

type ObjectContainer struct {
	metav1.TypeMeta `json:",inline"`
}

func NewCmdFuse() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "fuse-chart",
		Short:             `Fuse YAMLs`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			tplDir := filepath.Join(chartDir, chartName, "templates")
			err := os.MkdirAll(tplDir, 0755)
			if err != nil {
				return err
			}
			err = GenerateChartMetadata()
			if err != nil {
				return err
			}

			err = lib.ProcessDir(filepath.Join(sampleDir, chartName), func(obj *unstructured.Unstructured) error {
				rsKey, err := lib.ResourceKey(obj.GetAPIVersion(), obj.GetKind(), chartName, obj.GetName())
				if err != nil {
					return err
				}
				resourceKeys.Insert(rsKey)
				_, _, rsFilename := lib.ResourceFilename(obj.GetAPIVersion(), obj.GetKind(), chartName, obj.GetName())

				// values
				modelValues[rsKey] = ObjectContainer{
					TypeMeta: metav1.TypeMeta{
						APIVersion: obj.GetAPIVersion(),
						Kind:       obj.GetKind(),
					},
				}

				// schema
				gvr, err := registry.GVR(obj.GetObjectKind().GroupVersionKind())
				if err != nil {
					return err
				}
				descriptor, err := registry.LoadByGVR(gvr)
				if err != nil {
					return err
				}
				if descriptor.Spec.Validation != nil && descriptor.Spec.Validation.OpenAPIV3Schema != nil {
					delete(descriptor.Spec.Validation.OpenAPIV3Schema.Properties, "status")
					chartSchema.Properties[rsKey] = *descriptor.Spec.Validation.OpenAPIV3Schema
				}

				// templates
				filename := filepath.Join(tplDir, rsFilename+".yaml")
				f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close()

				objModel := ObjectModel{
					Key:    rsKey,
					Object: obj,
				}
				modelJSON, err := json.Marshal(objModel)
				if err != nil {
					return err
				}

				var data map[string]interface{}
				err = json.Unmarshal(modelJSON, &data)
				if err != nil {
					panic(err)
				}

				resourceTemplate := `{{"{{- with .Values."}}{{ .key }} {{"}}"}}
{{"{{- . | toYaml }}"}}
{{"{{- end }}"}}
`
				funcMap := sprig.TxtFuncMap()
				funcMap["toYaml"] = toYAML
				funcMap["toJson"] = toJSON
				tpl := template.Must(template.New("resourceTemplate").Funcs(funcMap).Parse(resourceTemplate))
				err = tpl.Execute(f, &data)
				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}

			{
				removeDescription(&chartSchema)
				data3, err := yaml.Marshal(chartSchema)
				if err != nil {
					return err
				}
				schemaFilename := filepath.Join(chartDir, chartName, "values.openapiv3_schema.yaml")
				err = ioutil.WriteFile(schemaFilename, data3, 0644)
				if err != nil {
					return err
				}
			}

			{
				data, err := yaml.Marshal(modelValues)
				if err != nil {
					panic(err)
				}

				var root y3.Node
				err = y3.Unmarshal(data, &root)
				if err != nil {
					return err
				}
				addDocComments(&root)

				//data, err = y3.Marshal(&root)
				//if err != nil {
				//	return err
				//}

				var buf bytes.Buffer
				enc := y3.NewEncoder(&buf)
				enc.SetIndent(2)
				defer enc.Close()
				err = enc.Encode(&root)
				if err != nil {
					return err
				}

				filename := filepath.Join(chartDir, chartName, "values.yaml")
				err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
				if err != nil {
					return err
				}
			}

			{
				desc := flect.Titleize(strings.ReplaceAll(chartName, "-", " "))
				doc := api.DocInfo{
					Project: api.ProjectInfo{
						Name:        fmt.Sprintf("%s by AppsCode", desc),
						ShortName:   fmt.Sprintf("%s", desc),
						URL:         "https://byte.builders",
						Description: fmt.Sprintf("%s", desc),
						App:         fmt.Sprintf("a %s", desc),
					},
					Repository: api.RepositoryInfo{
						URL:  "https://bundles.bytebuilders.dev/ui/",
						Name: "bytebuilders-ui",
					},
					Chart: api.ChartInfo{
						Name:          chartName,
						Version:       "v0.1.0",
						Values:        "-- generate from values file --",
						ValuesExample: "-- generate from values file --",
					},
					Prerequisites: []string{
						"Kubernetes 1.14+",
					},
					Release: api.ReleaseInfo{
						Name:      chartName,
						Namespace: metav1.NamespaceDefault,
					},
				}

				data, err := yaml.Marshal(&doc)
				if err != nil {
					return err
				}

				filename := filepath.Join(chartDir, chartName, "doc.yaml")
				err = ioutil.WriteFile(filename, data, 0644)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sampleDir, "sample-dir", sampleDir, "Sample dir")
	cmd.Flags().StringVar(&chartDir, "chart-dir", chartDir, "Charts dir")
	cmd.Flags().StringVar(&chartName, "chart-name", chartName, "Charts name")

	return cmd
}

func GenerateChartMetadata() error {
	chartMeta := chart.Metadata{
		Name:        chartName,
		Home:        "https://byte.builders",
		Version:     "v0.1.0",
		Description: "Ui Wizard Chart",
		Keywords:    []string{"appscode"},
		Maintainers: []*chart.Maintainer{
			{
				Name:  "AppsCode Engineering",
				Email: "support@appscode.com",
				URL:   "https://appscode.com",
			},
		},
	}
	data4, err := yaml.Marshal(chartMeta)
	if err != nil {
		return err
	}
	filename := filepath.Join(chartDir, chartName, "Chart.yaml")
	return ioutil.WriteFile(filename, data4, 0644)
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// toJSON takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

func addDocComments(node *y3.Node) {
	if node.Tag == "!!str" && resourceKeys.Has(node.Value) {
		node.LineComment = "# +doc-gen:break"
	}
	for i := range node.Content {
		addDocComments(node.Content[i])
	}
}

// removeDescription removes defaults from apiextensions.k8s.io/v1beta1 CRD definition.
func removeDescription(schema *crdv1.JSONSchemaProps) {
	if schema == nil {
		return
	}

	schema.Description = ""

	if schema.Items != nil {
		removeDescription(schema.Items.Schema)

		for idx := range schema.Items.JSONSchemas {
			removeDescription(&schema.Items.JSONSchemas[idx])
		}
	}

	for idx := range schema.AllOf {
		removeDescription(&schema.AllOf[idx])
	}
	for idx := range schema.OneOf {
		removeDescription(&schema.OneOf[idx])
	}
	for idx := range schema.AnyOf {
		removeDescription(&schema.AnyOf[idx])
	}
	if schema.Not != nil {
		removeDescription(schema.Not)
	}
	for key, prop := range schema.Properties {
		removeDescription(&prop)
		schema.Properties[key] = prop
	}
	if schema.AdditionalProperties != nil {
		removeDescription(schema.AdditionalProperties.Schema)
	}
	for key, prop := range schema.PatternProperties {
		removeDescription(&prop)
		schema.PatternProperties[key] = prop
	}
	for key, prop := range schema.Dependencies {
		removeDescription(prop.Schema)
		schema.Dependencies[key] = prop
	}
	if schema.AdditionalItems != nil {
		removeDescription(schema.AdditionalItems.Schema)
	}
	for key, prop := range schema.Definitions {
		removeDescription(&prop)
		schema.Definitions[key] = prop
	}
}
