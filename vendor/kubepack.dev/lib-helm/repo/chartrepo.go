/*
Copyright The Helm Authors.

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

package repo // import "helm.sh/helm/v3/pkg/repo"

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"
	"kubepack.dev/lib-helm/getter"
)

// Entry represents a collection of parameters for chart repository
type Entry struct {
	// Deprecated
	Name string `json:"name"`

	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	CAFile   string `json:"caFile"`

	Cache httpcache.Cache `json:"-"`
}

// ChartRepository represents a chart repository
type ChartRepository struct {
	Config     *Entry
	ChartPaths []string
	IndexFile  *IndexFile
	Client     getter.Getter
}

// NewChartRepository constructs ChartRepository
func NewChartRepository(cfg *Entry, getters getter.Providers) (*ChartRepository, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, errors.Errorf("invalid chart URL format: %s", cfg.URL)
	}

	client, err := getters.ByScheme(u.Scheme)
	if err != nil {
		return nil, errors.Errorf("could not find protocol handler for: %s", u.Scheme)
	}

	return &ChartRepository{
		Config:    cfg,
		IndexFile: NewIndexFile(),
		Client:    client,
	}, nil
}

// DownloadIndexFile fetches the index from a repository.
func (r *ChartRepository) DownloadIndexFile() (*bytes.Reader, error) {
	parsedURL, err := url.Parse(r.Config.URL)
	if err != nil {
		return nil, err
	}
	parsedURL.RawPath = path.Join(parsedURL.RawPath, "index.yaml")
	parsedURL.Path = path.Join(parsedURL.Path, "index.yaml")

	indexURL := parsedURL.String()
	// TODO add user-agent
	return r.Client.Get(indexURL,
		getter.WithURL(r.Config.URL),
		getter.WithTLSClientConfig(r.Config.CertFile, r.Config.KeyFile, r.Config.CAFile),
		getter.WithBasicAuth(r.Config.Username, r.Config.Password),
		getter.WithCache(r.Config.Cache),
	)
}

// Load loads the index for the chart repository
func (r *ChartRepository) Load() error {
	idx, err := r.DownloadIndexFile()
	if err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", r.Config.URL)
	}

	// Read the index file for the repository to get chart information and return chart URL
	repoIndex, err := LoadIndexFile(idx)
	if err != nil {
		return err
	}
	r.IndexFile = repoIndex

	return nil
}

func (r *ChartRepository) ListCharts() []string {
	charts := []string{}
	for name, varsions := range r.IndexFile.Entries {
		if varsions[0].Type == "" || varsions[0].Type == "application" {
			charts = append(charts, name)
		}
	}
	sort.Strings(charts)
	return charts
}

func (r *ChartRepository) ListVersions(name string) []VersionView {
	versions := []VersionView{}
	for _, v := range r.IndexFile.Entries[name] {
		if v.Removed {
			continue
		}
		versions = append(versions, VersionView{
			Version:     v.Version,
			AppVersion:  v.AppVersion,
			Deprecated:  v.Deprecated,
			KubeVersion: v.KubeVersion,
			Created:     v.Created,
			Removed:     v.Removed,
		})
	}
	return versions
}

// FindChartInAuthRepoURL finds chart in chart repository pointed by repoURL
// without adding repo to repositories, like FindChartInRepoURL,
// but it also receives credentials for the chart repository.
func FindChartInAuthRepoURL(rc *Entry, chartName, chartVersion string, getters getter.Providers) (*ChartVersion, error) {
	r, err := NewChartRepository(rc, getters)
	if err != nil {
		return nil, err
	}
	idx, err := r.DownloadIndexFile()
	if err != nil {
		return nil, errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", rc.URL)
	}

	// Read the index file for the repository to get chart information and return chart URL
	repoIndex, err := LoadIndexFile(idx)
	if err != nil {
		return nil, err
	}

	errMsg := fmt.Sprintf("chart %q", chartName)
	if chartVersion != "" {
		errMsg = fmt.Sprintf("%s version %q", errMsg, chartVersion)
	}
	cv, err := repoIndex.Get(chartName, chartVersion)
	if err != nil {
		return nil, errors.Errorf("%s not found in %s repository", errMsg, rc.URL)
	}

	if len(cv.URLs) == 0 {
		return nil, errors.Errorf("%s has no downloadable URLs", errMsg)
	}

	chartURL := cv.URLs[0]

	absoluteChartURL, err := ResolveReferenceURL(rc.URL, chartURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make chart URL absolute")
	}

	out := *cv
	out.URLs = []string{absoluteChartURL}
	return &out, nil
}

// ResolveReferenceURL resolves refURL relative to baseURL.
// If refURL is absolute, it simply returns refURL.
func ResolveReferenceURL(baseURL, refURL string) (string, error) {
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %s as URL", baseURL)
	}

	parsedRefURL, err := url.Parse(refURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %s as URL", refURL)
	}

	// We need a trailing slash for ResolveReference to work, but make sure there isn't already one
	parsedBaseURL.Path = strings.TrimSuffix(parsedBaseURL.Path, "/") + "/"
	return parsedBaseURL.ResolveReference(parsedRefURL).String(), nil
}

// VersionView represents a chart version entry in the IndexFile
type VersionView struct {
	// A SemVer 2 conformant version string of the chart
	Version string `json:"version,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `json:"deprecated,omitempty"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string    `json:"kubeVersion,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Removed     bool      `json:"removed,omitempty"`
}
