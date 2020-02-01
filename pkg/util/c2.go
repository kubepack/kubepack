package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"kubepack.dev/lib-helm/downloader"
	"kubepack.dev/lib-helm/getter"
	"kubepack.dev/lib-helm/repo"

	"helm.sh/helm/v3/pkg/provenance"
)

// LocateChart looks for a chart and returns either the reader or an error.
func LocateChart(reg *repo.Registry, repoURL, name, version string) (*bytes.Reader, error) {
	if repoURL == "" {
		return nil, fmt.Errorf("can't find repoURL for chart %s", name)
	}

	rc, _, err := reg.Get(repoURL)
	if err != nil {
		return nil, err
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

	cv, err := repo.FindChartInAuthRepoURL(rc, name, version, getter.All())
	if err != nil {
		return nil, err
	}

	reader, err := dl.DownloadTo(cv.URLs[0], version)
	if err != nil {
		return nil, err
	}

	digest, err := provenance.Digest(reader)
	if err != nil {
		return nil, err
	}

	if cv.Digest != "" && cv.Digest != digest {
		// Need to download
	}

	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return reader, nil
}
