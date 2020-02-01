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
	"os"
	"sync"

	"github.com/PuerkitoBio/purell"
	"github.com/gomodule/redigo/redis"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	rediscache "github.com/gregjones/httpcache/redis"
	"helm.sh/helm/v3/pkg/helmpath"
)

type Registry struct {
	repos map[string]*Entry
	cache httpcache.Cache
	m     sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{repos: make(map[string]*Entry)}
}

func NewMemoryCacheRegistry() *Registry {
	return &Registry{repos: make(map[string]*Entry), cache: httpcache.NewMemoryCache()}
}

func NewDiskCacheRegistry() *Registry {
	dir := helmpath.CachePath("kubepack")
	_ = os.MkdirAll(dir, 0755)
	return &Registry{repos: make(map[string]*Entry), cache: diskcache.New(dir)}
}

func NewRedisCacheRegistry(client redis.Conn) *Registry {
	return &Registry{repos: make(map[string]*Entry), cache: rediscache.NewWithClient(client)}
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
	entry.Cache = r.cache
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
