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

package downloader

import (
	"bytes"
	"io"
	"net/url"

	"github.com/pkg/errors"
	"kubepack.dev/lib-helm/pkg/getter"
)

// VerificationStrategy describes a strategy for determining whether to verify a chart.
type VerificationStrategy int

const (
	// VerifyNever will skip all verification of a chart.
	VerifyNever VerificationStrategy = iota
	// VerifyIfPossible will attempt a verification, it will not error if verification
	// data is missing. But it will not stop processing if verification fails.
	VerifyIfPossible
	// VerifyAlways will always attempt a verification, and will fail if the
	// verification fails.
	VerifyAlways
	// VerifyLater will fetch verification data, but not do any verification.
	// This is to accommodate the case where another step of the process will
	// perform verification.
	VerifyLater
)

// ErrNoOwnerRepo indicates that a given chart URL can't be found in any repos.
var ErrNoOwnerRepo = errors.New("could not find a repo containing the given URL")

// ChartDownloader handles downloading a chart.
//
// It is capable of performing verifications on charts as well.
type ChartDownloader struct {
	// Out is the location to write warning and info messages.
	Out io.Writer
	// Getter collection for the operation
	Getters getter.Providers
	// Options provide parameters to be passed along to the Getter being initialized.
	Options []getter.Option
}

// DownloadTo retrieves a chart. Depending on the settings, it may also download a provenance file.
//
// Returns a string path to the location where the file was downloaded or an error if something bad happened.
func (c *ChartDownloader) DownloadTo(ref, version string) (*bytes.Reader, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return nil, errors.Errorf("invalid chart URL format: %s", ref)
	}

	g, err := c.Getters.ByScheme(u.Scheme)
	if err != nil {
		return nil, err
	}

	return g.Get(u.String(), c.Options...)
}
