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

package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/helmpath"
	"kubepack.dev/lib-helm/pkg/chart/loader"
	"kubepack.dev/lib-helm/pkg/downloader"
	"kubepack.dev/lib-helm/pkg/getter"
)

type Registry struct {
	repos map[string]*Entry
	cache httpcache.Cache
	m     sync.RWMutex
}

var _ IRegistry = &Registry{}

func NewCachedRegistry(cache httpcache.Cache) *Registry {
	return &Registry{repos: make(map[string]*Entry), cache: cache}
}

func NewRegistry() *Registry {
	return NewCachedRegistry(nil)
}

func NewMemoryCacheRegistry() *Registry {
	return NewCachedRegistry(httpcache.NewMemoryCache())
}

func NewDiskCacheRegistry() *Registry {
	dir := helmpath.CachePath("kubepack")
	_ = os.MkdirAll(dir, 0o755)
	return NewCachedRegistry(diskcache.New(dir))
}

func (r *Registry) Add(e *Entry) error {
	url, err := purell.NormalizeURLString(e.URL, purell.FlagsUsuallySafeGreedy)
	if err != nil {
		return err
	}
	e.URL = url

	r.m.Lock()
	r.repos[url] = e
	r.m.Unlock()

	return nil
}

func (r *Registry) Get(url string) (*Entry, bool, error) {
	url, err := purell.NormalizeURLString(url, purell.FlagsUsuallySafeGreedy)
	if err != nil {
		return nil, false, err
	}

	r.m.RLock()
	entry, ok := r.repos[url]
	if !ok {
		entry = &Entry{
			URL: url,
		}
	}
	if entry.Username != "" || entry.Password != "" || entry.CAFile != "" || entry.CertFile != "" || entry.KeyFile != "" {
		entry.Cache = nil
	} else {
		entry.Cache = r.cache
	}
	r.m.RUnlock()

	return entry, ok, nil
}

func (r *Registry) Delete(url string) (*Entry, error) {
	url, err := purell.NormalizeURLString(url, purell.FlagsUsuallySafeGreedy)
	if err != nil {
		return nil, err
	}

	r.m.Lock()
	entry := r.repos[url]
	delete(r.repos, url)
	r.m.Unlock()

	return entry, nil
}

// LocateChart looks for a chart and returns either the reader or an error.
func (r *Registry) LocateChart(repository, name, version string) (loader.ChartLoader, *ChartVersion, error) {
	repository = strings.TrimSpace(repository)
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)

	if dir, ok := os.LookupEnv("UI_WIZARD_CHARTS_DIR"); ok {
		repository = filepath.Join(dir, name)
	}

	if repository == "" {
		return nil, nil, fmt.Errorf("can't find repository for chart %s", name)
	}

	if fi, err := os.Stat(repository); err == nil {
		abs, err := filepath.Abs(repository)
		if err != nil {
			return nil, nil, err
		}
		//if c.Verify {
		//	if _, err := downloader.VerifyChart(abs, c.Keyring); err != nil {
		//		return "", err
		//	}
		//}
		if fi.IsDir() {
			return loader.DirLoader(abs), nil, nil
		}
		return loader.FileLoader(abs), nil, nil
	}
	if filepath.IsAbs(repository) || strings.HasPrefix(repository, ".") {
		return nil, nil, errors.Errorf("path %q not found", repository)
	}

	rc, _, err := r.Get(repository)
	if err != nil {
		return nil, nil, err
	}

	dl := downloader.ChartDownloader{
		Out:     os.Stdout,
		Getters: getter.All(),
		Options: []getter.Option{
			getter.WithURL(rc.URL),
			getter.WithTLSClientConfig(rc.CertFile, rc.KeyFile, rc.CAFile),
			getter.WithBasicAuth(rc.Username, rc.Password),
			getter.WithCache(rc.Cache),
		},
	}

	cv, err := FindChartInAuthRepoURL(rc, name, version, getter.All())
	if err != nil {
		return nil, nil, err
	}

	reader, err := dl.DownloadTo(cv.URLs[0], version)
	if err != nil {
		return nil, nil, err
	}

	// digest, err := provenance.Digest(reader)
	// if err != nil {
	// 	return nil, err
	// }

	// if cv.Digest != "" && cv.Digest != digest {
	// 	// Need to download
	// }

	l2 := loader.ByteLoader(*reader)
	return &l2, cv, nil
}

func (r *Registry) GetChart(repository, chartName, chartVersion string) (*ChartExtended, error) {
	chartLoader, cv, err := r.LocateChart(repository, chartName, chartVersion)
	if err != nil {
		return nil, err
	}

	chrt, err := chartLoader.Load()
	if err != nil {
		return nil, err
	}

	cx := &ChartExtended{
		Chart: chrt,
	}
	if cv != nil {
		cx.URLs = cv.URLs
		cx.Created = cv.Created
		cx.Removed = cv.Removed
		cx.Digest = cv.Digest
	}
	return cx, nil
}

// ChartExtended represents a chart with metadata from its entry in the IndexFile
type ChartExtended struct {
	*chart.Chart
	URLs    []string  `json:"urls"`
	Created time.Time `json:"created,omitempty"`
	Removed bool      `json:"removed,omitempty"`
	Digest  string    `json:"digest,omitempty"`
}
