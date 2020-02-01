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
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
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
	m      sync.Mutex
}

func NewCachedRepoNamer() *RepoNamer {
	return &RepoNamer{
		client: &http.Client{
			Transport: httpcache.NewMemoryCacheTransport(),
		},
		repos: make(map[string]string),
	}
}

func (r *RepoNamer) getHelmHubAlias(chartURL string) (string, bool) {
	resp, err := r.client.Get("https://raw.githubusercontent.com/helm/hub/master/repos.yaml")
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}

	var hub Hub
	err = yaml.Unmarshal(data, &hub)
	if err != nil {
		return "", false
	}

	r.m.Lock()
	for _, repo := range hub.Repositories {
		u, err := purell.NormalizeURLString(repo.URL, purell.FlagsUsuallySafeGreedy)
		if err != nil {
			return "", false
		}
		r.repos[u] = repo.Name
	}
	r.m.Unlock()

	name, ok := r.repos[chartURL]
	return name, ok
}

func (r *RepoNamer) Name(chartURL string) (string, error) {
	var err error
	chartURL, err = purell.NormalizeURLString(chartURL, purell.FlagsUsuallySafeGreedy)
	if err != nil {
		return "", err
	}

	r.m.Lock()
	name, ok := r.repos[chartURL]
	r.m.Unlock()
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
		r.m.Lock()
		r.repos[chartURL] = name
		r.m.Unlock()
		return name, nil
	} else if ipv4 := ip.To4(); ipv4 != nil {
		return strings.ReplaceAll(ipv4.String(), ".", "-"), nil
	} else if ipv6 := ip.To16(); ipv6 != nil {
		return strings.ReplaceAll(ipv6.String(), ":", "-"), nil
	}
	return "", fmt.Errorf("failed to generate repo name for url:%s", chartURL)
}

func (r *RepoNamer) nameForDomain(domain string) string {
	// TODO: Use https://raw.githubusercontent.com/helm/hub/master/repos.yaml
	if strings.HasSuffix(domain, ".storage.googleapis.com") {
		return strings.TrimSuffix(domain, ".storage.googleapis.com")
	}

	publicSuffix, icann := publicsuffix.PublicSuffix(domain)
	if icann {
		domain = strings.TrimSuffix(domain, "."+publicSuffix)
	}
	if strings.HasPrefix(domain, "charts.") {
		domain = strings.TrimPrefix(domain, "charts.")
	}

	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts)/2; i++ {
		j := len(parts) - i - 1
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "-")
}
