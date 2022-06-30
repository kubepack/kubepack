package chart

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
)

// helm.sh/helm/v3/pkg/action/install.go
const notesFileSuffix = "NOTES.txt"

// RenderResources renders the templates in a chart
func RenderResources(ch *chart.Chart, caps *chartutil.Capabilities, values chartutil.Values) ([]*release.Hook, []releaseutil.Manifest, error) {
	hs := []*release.Hook{}
	b := bytes.NewBuffer(nil)

	if ch.Metadata.KubeVersion != "" {
		if !chartutil.IsCompatibleRange(ch.Metadata.KubeVersion, caps.KubeVersion.String()) {
			return hs, nil, errors.Errorf("chart requires kubeVersion: %s which is incompatible with Kubernetes %s", ch.Metadata.KubeVersion, caps.KubeVersion.String())
		}
	}

	files, err := engine.Render(ch, values)
	if err != nil {
		return hs, nil, err
	}

	for k := range files {
		if strings.HasSuffix(k, notesFileSuffix) {
			delete(files, k)
		}
	}

	// Sort hooks, manifests, and partials. Only hooks and manifests are returned,
	// as partials are not used after renderer.Render. Empty manifests are also
	// removed here.
	hs, manifests, err := releaseutil.SortManifests(files, caps.APIVersions, releaseutil.InstallOrder)
	if err != nil {
		// By catching parse errors here, we can prevent bogus releases from going
		// to Kubernetes.
		//
		// We return the files as a big blob of data to help the user debug parser
		// errors.
		for name, content := range files {
			if strings.TrimSpace(content) == "" {
				continue
			}
			fmt.Fprintf(b, "---\n# Source: %s\n%s\n", name, content)
		}
		return hs, manifests, err
	}

	return hs, manifests, nil
}

func IsEvent(events []release.HookEvent, x release.HookEvent) bool {
	for _, event := range events {
		if event == x {
			return true
		}
	}
	return false
}

// IsChartInstallable validates if a chart can be installed
//
// Application chart type is only installable
func IsChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
