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
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/helmpath"
	"kubepack.dev/lib-helm/chart/loader"
	"kubepack.dev/lib-helm/downloader"
	"kubepack.dev/lib-helm/getter"
)

type Registry struct {
	repos map[string]*Entry
	cache httpcache.Cache
	m     sync.RWMutex
}

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
	_ = os.MkdirAll(dir, 0755)
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
func (r *Registry) LocateChart(repoURL, name, version string) (*bytes.Reader, *ChartVersion, error) {
	if repoURL == "" {
		return nil, nil, fmt.Errorf("can't find repoURL for chart %s", name)
	}

	rc, _, err := r.Get(repoURL)
	if err != nil {
		return nil, nil, err
	}

	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)

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

	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}

	return reader, cv, nil
}

func (r *Registry) GetChart(repoURL, chartName, chartVersion string) (*ChartExtended, error) {
	reader, cv, err := r.LocateChart(repoURL, chartName, chartVersion)
	if err != nil {
		return nil, err
	}

	chrt, err := loader.Load(reader)
	if err != nil {
		return nil, err
	}

	return &ChartExtended{
		Chart:   chrt,
		URLs:    cv.URLs,
		Created: cv.Created,
		Removed: cv.Removed,
		Digest:  cv.Digest,
	}, nil
}

// ChartExtended represents a chart with metadata from its entry in the IndexFile
type ChartExtended struct {
	*chart.Chart
	URLs    []string  `json:"urls"`
	Created time.Time `json:"created,omitempty"`
	Removed bool      `json:"removed,omitempty"`
	Digest  string    `json:"digest,omitempty"`
}
