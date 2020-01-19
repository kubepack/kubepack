/*
Copyright The Kmodules Authors.

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

package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	pkgvalues "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func InstallChart(chartName, repoName, url, name, namespace string, values []string, factory genericclioptions.RESTClientGetter) error {
	_, err := getChart(chartName, repoName, url, name, namespace, values, factory, true)
	return err
}

func getChart(chartName, repoName, url, name, namespace string, values []string, factory genericclioptions.RESTClientGetter, install bool) ([]string, error) {
	cfg := new(action.Configuration)

	chartDir, err := ioutil.TempDir(TmpDir, DirPrefix)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(chartDir)

	err = setEnv(chartDir)
	if err != nil {
		return nil, err
	}
	defer unsetEnv()

	settings := cli.New()

	err = getRepo(repoName, url)
	if err != nil {
		return nil, err
	}

	if namespace == "" {
		namespace = settings.Namespace()
	}

	if install {
		if err := cfg.Init(factory, namespace, "", debug); err != nil {
			return nil, err
		}
	}

	client := action.NewInstall(cfg)
	if name == "" {
		client.GenerateName = true
		var err error

		name, _, err = client.NameAndChart([]string{chartName})
		if err != nil {
			return nil, err
		}
	}

	client.ReleaseName = name
	client.Namespace = namespace

	cp, err := client.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	valueOpts := pkgvalues.Options{
		Values: values,
	}

	p := getter.All(settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	client.DryRun = !install
	client.ClientOnly = !install

	rel, err := client.Run(chartRequested, vals)
	if err != nil {
		return nil, err
	}

	var out []string
	splitYAML([]byte(rel.Manifest), &out)

	return out, nil
}

func GetChangedValues(original map[string]interface{}, modified map[string]interface{}) []string {
	var cmd []string
	getChangedValues(original, modified, "", &cmd)
	return cmd
}

func getChangedValues(original map[string]interface{}, modified map[string]interface{}, currentKey string, setCmds *[]string) {
	for key, val := range original {
		tempKey := ""
		if currentKey == "" {
			tempKey = key
		} else {
			tempKey = currentKey + "." + key
		}

		switch val := val.(type) {
		case map[string]interface{}:
			getChangedValues(val, modified[key].(map[string]interface{}), tempKey, setCmds)
		case []interface{}:
			if !cmp.Equal(val, modified[key]) {
				tempCmd := tempKey + "=["
				arrayLen := len(modified[key].([]interface{}))
				for i, element := range modified[key].([]interface{}) {
					tempCmd = tempCmd + fmt.Sprintf("%v", element)
					if i != arrayLen-1 {
						tempCmd = tempCmd + fmt.Sprintf(",")
					}
				}
				tempCmd = tempCmd + "]"
				if strings.Contains(tempCmd, "=[]") {
					tempCmd = strings.ReplaceAll(tempCmd, "[]", "null")
				}
				*setCmds = append(*setCmds, tempCmd)
			}
		case interface{}:
			if val != modified[key] {
				if isZeroOrNil(modified[key]) {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=null ", tempKey))
				} else {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=%v ", tempKey, modified[key]))
				}
			}
		default:
			if val != modified[key] {
				if isZeroOrNil(modified[key]) {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=null ", tempKey))
				} else {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=%v ", tempKey, modified[key]))
				}
			}
		}
	}
}

func isZeroOrNil(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
