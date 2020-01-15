/*
Copyright The Kubepack Authors.

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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath/xdg"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	TmpDir    = "/tmp"
	DirPrefix = "helm"
)

var name = "stash-bundle"
var charts = []string{
	"https://charts.appscode.com/stable/@stash@v0.9.0-rc.2",
}
var addons = []string{}

/*
$ go run cmd/bundle-generator/main.go --name=stash-bundle --charts https://charts.appscode.com/stable/@stash@v0.9.0-rc.2

$ go run cmd/bundle-generator/main.go --name=kubedb \
  --charts https://charts.appscode.com/stable/@kubedb@v0.9.0-rc.2 \
  --charts https://charts.appscode.com/stable/@kubedb-catalog@v0.9.0-rc.2
  --addons https://kubepack-testcharts.storage.googleapis.com@stash-bundle@v0.1.0
*/
func main() {
	flag.StringVar(&name, "name", name, "Name of bundle, example: stash-bundle")
	flag.StringArrayVar(&charts, "charts", charts, "Provide list charts in this bundle. format --charts url=chart_name@version --charts url=chart_name@version")
	flag.StringArrayVar(&addons, "addons", addons, "Provide list of addons in this bundle. format --addons url=bundle_chart_name@v1,v2 --addons url=bundle_chart_name@v1,v2")
	flag.Parse()

	fmt.Println(charts)

	/*
	  name:
	  labels:
	    {{- include "stash-mongodb-bundle.labels" . : nindent 4 }}
	*/
	b := v1alpha1.Bundle{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindBundle,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf(`{{ include "%s.fullname" . }}`, name),
		},
	}

	for _, val := range charts {
		parts := strings.Split(val, "@")
		url := parts[0]
		chartName := parts[1]
		versions := strings.Split(parts[2], ",")
		primaryVersion := strings.SplitN(versions[0], ":", 2)[0]
		required := true
		if len(parts) >= 4 && strings.EqualFold(parts[3], "optional") {
			required = false
		}
		multiSelect := false
		if len(parts) >= 5 && strings.EqualFold(parts[4], "anyof") {
			multiSelect = true
		}

		pkgChart, err := GetChart(chartName, primaryVersion, "myrepo", url)
		if err != nil {
			log.Fatalln(err)
		}
		ref := v1alpha1.PackageRef{
			Chart: v1alpha1.ChartOption{
				ChartRef: v1alpha1.ChartRef{
					URL:     url,
					Name:    pkgChart.Name(),
					Feature: pkgChart.Metadata.Description,
				},
			},
			Required: required,
		}
		for _, versionInfo := range versions {
			vparts := strings.SplitN(versionInfo, ":", 2)
			version := vparts[0]
			selected := false
			if len(vparts) == 2 {
				selected, _ = strconv.ParseBool(vparts[1])
			}
			ref.Chart.Versions = append(ref.Chart.Versions, v1alpha1.VersionOption{
				Version:  version,
				Selected: selected,
			})
			ref.Chart.MultiSelect = multiSelect
		}
		b.Spec.Packages = append(b.Spec.Packages, ref)
	}

	for _, val := range addons {
		parts := strings.SplitN(val, "@", 3)
		url := parts[0]
		bundleName := parts[1]
		versions := strings.Split(parts[2], ",")

		chart, err := GetChart(bundleName, versions[0], "myrepo", url)
		if err != nil {
			log.Fatalln(err)
		}
		addon := v1alpha1.Addon{
			Feature: chart.Metadata.Description,
			Bundle: &v1alpha1.BundleOption{
				BundleRef: v1alpha1.BundleRef{
					URL:  url,
					Name: chart.Name(),
				},
			},
		}
		for idx, v := range versions {
			addon.Bundle.Versions = append(addon.Bundle.Versions, v1alpha1.VersionOption{
				Version:  v,
				Selected: idx == 0,
			})
		}
		b.Spec.Addons = append(b.Spec.Addons, addon)
	}

	data, err := yaml.Marshal(b)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("testdata/charts/"+name+"/templates", 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("testdata/charts/"+name+"/templates/bundle.yaml", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func GetChart(chartName, version, repoName, url string) (*chart.Chart, error) {
	cfg := new(action.Configuration)
	client := action.NewInstall(cfg)
	client.Version = version
	client.RepoURL = url

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
	//err = getRepo(repoName, url)
	//if err != nil {
	//	return nil, err
	//}

	cp, err := client.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	return chartRequested, nil
}

func getRepo(name string, url string) error {
	settings := cli.New()

	repofile := settings.RepositoryConfig
	err := os.MkdirAll(filepath.Dir(repofile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	fileLock := flock.New(strings.Replace(repofile, filepath.Ext(repofile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(repofile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	if f.Has(name) {
		return errors.Errorf("repository name (%s) already exists, please specify a different name", name)
	}

	c := repo.Entry{
		Name: name,
		URL:  url,
	}
	f.Update(&c)

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if err := f.WriteFile(repofile, 0644); err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
	}

	return nil
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	log.Output(2, fmt.Sprintf(format, v...))
}

func setEnv(chartDir string) error {
	err := os.Setenv(xdg.CacheHomeEnvVar, filepath.Join(chartDir, "cache"))
	if err != nil {
		return err
	}

	err = os.Setenv(xdg.ConfigHomeEnvVar, filepath.Join(chartDir, "config"))
	if err != nil {
		return err
	}

	return os.Setenv(xdg.DataHomeEnvVar, filepath.Join(chartDir, "data"))
}

func unsetEnv() error {
	err := os.Unsetenv(xdg.CacheHomeEnvVar)
	if err != nil {
		return err
	}

	err = os.Unsetenv(xdg.ConfigHomeEnvVar)
	if err != nil {
		return err
	}

	return os.Unsetenv(xdg.DataHomeEnvVar)
}
