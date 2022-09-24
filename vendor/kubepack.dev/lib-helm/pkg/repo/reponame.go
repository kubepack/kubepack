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
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/PuerkitoBio/purell"
	"github.com/gregjones/httpcache"
	"golang.org/x/net/publicsuffix"
	"sigs.k8s.io/yaml"
)

var DefaultNamer = NewCachedRepoNamer()

type Hub struct {
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type RepoNamer struct {
	client *http.Client
	repos  map[string]string // url -> name
	loaded bool
	m      sync.RWMutex
}

func NewCachedRepoNamer() *RepoNamer {
	return &RepoNamer{
		client: &http.Client{
			Transport: httpcache.NewMemoryCacheTransport(),
		},
		repos: make(map[string]string),
	}
}

func (r *RepoNamer) setAlias(url, alias string) {
	r.m.Lock()
	r.repos[url] = alias
	r.m.Unlock()
}

func (r *RepoNamer) getAlias(url string) (string, bool) {
	r.m.RLock()
	alias, ok := r.repos[url]
	r.m.RUnlock()
	return alias, ok
}

func (r *RepoNamer) listRepos() []Repository {
	r.m.RLock()

	repos := make([]Repository, 0, len(r.repos))
	for u, name := range r.repos {
		repos = append(repos, Repository{
			Name: name,
			URL:  u,
		})
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })

	defer r.m.RUnlock()
	return repos
}

func (r *RepoNamer) helmHubAliasesLoaded() bool {
	r.m.RLock()
	defer r.m.RUnlock()

	return r.loaded
}

func (r *RepoNamer) loadHelmHubAliases() error {
	if r.helmHubAliasesLoaded() {
		return nil
	}

	resp, err := r.client.Get("https://raw.githubusercontent.com/helm/hub/master/repos.yaml")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var hub Hub
	err = yaml.Unmarshal(data, &hub)
	if err != nil {
		return err
	}

	newRepos := make(map[string]string, len(hub.Repositories))
	r.m.RLock()
	for _, repo := range hub.Repositories {
		u, err := purell.NormalizeURLString(repo.URL, purell.FlagsUsuallySafeGreedy)
		if err != nil {
			return err
		}
		if _, found := r.repos[u]; !found {
			newRepos[u] = repo.Name
		}
	}
	r.m.RUnlock()

	if len(newRepos) > 0 {
		r.m.Lock()

		for u, alias := range newRepos {
			r.repos[u] = alias
		}
		r.loaded = true

		r.m.Unlock()
	}

	return nil
}

func (r *RepoNamer) ListHelmHubRepositories() ([]Repository, error) {
	err := r.loadHelmHubAliases()
	if err != nil {
		return nil, err
	}

	return r.listRepos(), nil
}

func (r *RepoNamer) getHelmHubAlias(chartURL string) (string, bool) {
	err := r.loadHelmHubAliases()
	if err != nil {
		return "", false
	}
	return r.getAlias(chartURL)
}

func (r *RepoNamer) Name(chartURL string) (string, error) {
	var err error
	chartURL, err = purell.NormalizeURLString(chartURL, purell.FlagsUsuallySafeGreedy)
	if err != nil {
		return "", err
	}

	name, ok := r.getAlias(chartURL)
	if ok {
		return name, nil
	}

	name, ok = r.getHelmHubAlias(chartURL)
	if ok {
		return name, nil
	}

	u, err := url.Parse(chartURL)
	if err != nil {
		return "", err
	}

	hostname := u.Hostname()
	ip := net.ParseIP(hostname)
	if ip == nil {
		name = r.nameForDomain(hostname)
		r.setAlias(chartURL, name)
		return name, nil
	} else if ipv4 := ip.To4(); ipv4 != nil {
		return strings.ReplaceAll(ipv4.String(), ".", "-"), nil
	} else if ipv6 := ip.To16(); ipv6 != nil {
		return strings.ReplaceAll(ipv6.String(), ":", "-"), nil
	}
	return "", fmt.Errorf("failed to generate repo name for url:%s", chartURL)
}

func (r *RepoNamer) nameForDomain(domain string) string {
	if domain == "kubernetes-charts.storage.googleapis.com" {
		return "stable"
	}
	if strings.HasSuffix(domain, ".storage.googleapis.com") {
		return strings.TrimSuffix(domain, ".storage.googleapis.com")
	}

	publicSuffix, icann := publicsuffix.PublicSuffix(domain)
	if icann {
		domain = strings.TrimSuffix(domain, "."+publicSuffix)
	}

	domain = strings.TrimPrefix(domain, "charts.")

	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts)/2; i++ {
		j := len(parts) - i - 1
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "-")
}
