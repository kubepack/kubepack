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
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/purell"
	fluxsrc "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"gomodules.xyz/sets"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/helmpath"
	core "k8s.io/api/core/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kubepack.dev/lib-helm/pkg/chart/loader"
	"kubepack.dev/lib-helm/pkg/downloader"
	"kubepack.dev/lib-helm/pkg/getter"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Registry struct {
	repos     map[string]*Entry
	kc        client.Reader
	helmrepos map[string]client.ObjectKey // url -> secret
	cache     httpcache.Cache
	m         sync.RWMutex
}

var _ IRegistry = &Registry{}

func NewCachedRegistry(kc client.Reader, cache httpcache.Cache) *Registry {
	reg := &Registry{
		repos:     make(map[string]*Entry),
		kc:        kc,
		helmrepos: make(map[string]client.ObjectKey),
		cache:     cache,
	}
	if kc != nil {
		var list fluxsrc.HelmRepositoryList
		if err := kc.List(context.TODO(), &list); err == nil {
			for _, item := range list.Items {
				if item.Spec.SecretRef != nil {
					if url, err := purell.NormalizeURLString(item.Spec.URL, purell.FlagsUsuallySafeGreedy); err == nil {
						reg.helmrepos[url] = client.ObjectKey{
							Namespace: item.Namespace,
							Name:      item.Spec.SecretRef.Name,
						}
					}
				}
			}
		}
	}
	return reg
}

func NewRegistry() *Registry {
	return NewCachedRegistry(nil, nil)
}

func NewMemoryCacheRegistry() *Registry {
	return NewCachedRegistry(nil, httpcache.NewMemoryCache())
}

func DefaultDiskCache() httpcache.Cache {
	dir := helmpath.CachePath("kubepack")
	_ = os.MkdirAll(dir, 0o755)
	return diskcache.New(dir)
}

func NewDiskCacheRegistry() *Registry {
	return NewCachedRegistry(nil, DefaultDiskCache())
}

func (r *Registry) Add(e *Entry) error {
	url, err := purell.NormalizeURLString(e.URL, purell.FlagsUsuallySafeGreedy)
	if err != nil {
		return err
	}
	e.URL = url
	if e.Username != "" || e.Password != "" || e.CAFile != "" || e.CertFile != "" || e.KeyFile != "" {
		e.Cache = nil
	} else {
		e.Cache = r.cache
	}

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
		if secretRef, hasAuth := r.helmrepos[url]; hasAuth {
			if err := r.addAuthInfo(secretRef, entry); err != nil {
				return nil, false, err
			}
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

func (r *Registry) Register(srcRef kmapi.TypedObjectReference) (string, error) {
	var repository string

	if srcRef.APIGroup == fluxsrc.GroupVersion.Group && srcRef.Kind == "HelmRepository" {
		if srcRef.Namespace == "" || srcRef.Name == "" {
			return "", fmt.Errorf("missing name or namespace for HelmRepository %+v", srcRef)
		}
		if r.kc == nil {
			return "", fmt.Errorf("kubernetes client not initialized for HelmRepository %+v", srcRef)
		}

		var src fluxsrc.HelmRepository
		err := r.kc.Get(context.TODO(), client.ObjectKey{Namespace: srcRef.Namespace, Name: srcRef.Name}, &src)
		if err != nil {
			return "", err
		}
		entry, found, err := r.Get(src.Spec.URL)
		if err != nil {
			return "", err
		}
		if !found {
			if src.Spec.SecretRef != nil {
				key := client.ObjectKey{Namespace: srcRef.Namespace, Name: src.Spec.SecretRef.Name}
				if err := r.addAuthInfo(key, entry); err != nil {
					return "", err
				}
			}

			// TODO(tamal): enforce
			// PassCredentials
			// src.Spec.AccessFrom

			if err = r.Add(entry); err != nil {
				return "", err
			}
		}
		repository = entry.URL
	} else {
		if srcRef.APIGroup != "" || srcRef.Kind != "" || srcRef.Namespace != "" {
			return "", fmt.Errorf("only repository name is expected, found %+v", srcRef)
		}
		repository = strings.TrimSpace(srcRef.Name)
	}
	return repository, nil
}

func (r *Registry) addAuthInfo(key client.ObjectKey, entry *Entry) error {
	var secret core.Secret
	err := r.kc.Get(context.TODO(), key, &secret)
	if err != nil {
		return err
	}
	if v, ok := secret.Data[core.BasicAuthUsernameKey]; ok {
		entry.Username = string(v)
	}
	if v, ok := secret.Data[core.BasicAuthPasswordKey]; ok {
		entry.Password = string(v)
	}

	// TODO(tamal): correct keys?
	if v, ok := secret.Data["ca.crt"]; ok {
		entry.CAFile = string(v)
	}
	if v, ok := secret.Data[core.TLSCertKey]; ok {
		entry.CertFile = string(v)
	}
	if v, ok := secret.Data[core.TLSPrivateKeyKey]; ok {
		entry.KeyFile = string(v)
	}
	return nil
}

var (
	bypassChartRegistries  []string
	once                   sync.Once
	bypassChartRegistrySet sets.String
)

// AddBypassChartRegistriesFlag is for explicitly initializing the --bypass-chart-registries flag
func AddBypassChartRegistriesFlag(fs *pflag.FlagSet) {
	if fs == nil {
		fs = pflag.CommandLine
	}
	fs.StringSliceVar(&bypassChartRegistries, "bypass-chart-registries", bypassChartRegistries, "List of Helm chart registries that can bypass using UI_WIZARD_CHARTS_DIR env variable")
}

// LocateChart looks for a chart and returns either the reader or an error.
func (r *Registry) LocateChart(repository, name, version string) (loader.ChartLoader, *ChartVersion, error) {
	repository = strings.TrimSpace(repository)
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)

	if repository == "" {
		return nil, nil, fmt.Errorf("can't find repository for chart %s", name)
	}

	once.Do(func() {
		bypassChartRegistrySet = sets.NewString(bypassChartRegistries...)
		bypassChartRegistrySet.Insert("charts.appscode.com", "bundles.byte.builders")
	})

	if u, err := url.Parse(repository); err == nil && bypassChartRegistrySet.Has(u.Hostname()) {
		if dir, ok := os.LookupEnv("UI_WIZARD_CHARTS_DIR"); ok {
			repository = filepath.Join(dir, name)
		}
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
