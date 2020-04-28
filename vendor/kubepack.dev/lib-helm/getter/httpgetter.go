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

package getter

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"kubepack.dev/lib-helm/internal/tlsutil"
	"kubepack.dev/lib-helm/internal/urlutil"
	"kubepack.dev/lib-helm/internal/version"
)

// HTTPGetter is the efault HTTP(/S) backend handler
type HTTPGetter struct {
	opts options
}

//Get performs a Get from repo.Getter and returns the body.
func (g *HTTPGetter) Get(href string, options ...Option) (*bytes.Reader, error) {
	for _, opt := range options {
		opt(&g.opts)
	}
	return g.get(href)
}

func (g *HTTPGetter) get(href string) (out *bytes.Reader, err error) {
	// Set a helm specific user agent so that a repo server and metrics can
	// separate helm calls from other tools interacting with repos.
	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", version.GetUserAgent())
	if g.opts.userAgent != "" {
		req.Header.Set("User-Agent", g.opts.userAgent)
	}

	if g.opts.username != "" && g.opts.password != "" {
		req.SetBasicAuth(g.opts.username, g.opts.password)
	}

	client, err := g.httpClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		e2 := resp.Body.Close()
		err = utilerrors.NewAggregate([]error{err, e2})
	}()
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("failed to fetch %s : %s", href, resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), err
}

// NewHTTPGetter constructs a valid http/https client as a Getter
func NewHTTPGetter(options ...Option) (Getter, error) {
	var client HTTPGetter

	for _, opt := range options {
		opt(&client.opts)
	}

	return &client, nil
}

func (g *HTTPGetter) httpClient() (*http.Client, error) {
	if (g.opts.certFile != "" && g.opts.keyFile != "") || g.opts.caFile != "" {
		tlsConf, err := tlsutil.NewClientTLS(g.opts.certFile, g.opts.keyFile, g.opts.caFile)
		if err != nil {
			return nil, errors.Wrap(err, "can't create TLS config for client")
		}
		// tlsConf.BuildNameToCertificate()

		sni, err := urlutil.ExtractHostname(g.opts.url)
		if err != nil {
			return nil, err
		}
		tlsConf.ServerName = sni

		// rt = http.DefaultTransport
		var rt http.RoundTripper = &http.Transport{
			TLSClientConfig: tlsConf,
			Proxy:           http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		if g.opts.cache != nil {
			rt = &httpcache.Transport{Cache: g.opts.cache, MarkCachedResponses: true, Transport: rt}
		}

		return &http.Client{
			Transport: rt,
		}, nil
	}

	if g.opts.cache != nil {
		return &http.Client{
			Transport: httpcache.NewTransport(g.opts.cache),
		}, nil
	}
	return http.DefaultClient, nil
}
