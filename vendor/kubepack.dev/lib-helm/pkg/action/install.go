package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ha "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"kmodules.xyz/resource-metadata/hub"
	libchart "kubepack.dev/lib-helm/pkg/chart"
	"kubepack.dev/lib-helm/pkg/flowapi"
	"kubepack.dev/lib-helm/pkg/repo"
	"kubepack.dev/lib-helm/pkg/values"
)

type InstallOptions struct {
	ChartURL                 string         `json:"chartURL"`
	ChartName                string         `json:"chartName"`
	Version                  string         `json:"version"`
	Values                   values.Options `json:",inline,omitempty"`
	DryRun                   bool           `json:"dryRun"`
	DisableHooks             bool           `json:"disableHooks"`
	Replace                  bool           `json:"replace"`
	Wait                     bool           `json:"wait"`
	Devel                    bool           `json:"devel"`
	Timeout                  time.Duration  `json:"timeout"`
	Namespace                string         `json:"namespace"`
	ReleaseName              string         `json:"releaseName"`
	Description              string         `json:"description"`
	Atomic                   bool           `json:"atomic"`
	SkipCRDs                 bool           `json:"skipCRDs"`
	SubNotes                 bool           `json:"subNotes"`
	DisableOpenAPIValidation bool           `json:"disableOpenAPIValidation"`
	IncludeCRDs              bool           `json:"includeCRDs"`
	PartOf                   string         `json:"partOf"`
}

type Installer struct {
	cfg *Configuration

	opts   InstallOptions
	reg    *repo.Registry
	result *release.Release
}

func NewInstaller(getter genericclioptions.RESTClientGetter, namespace string, helmDriver string) (*Installer, error) {
	cfg := new(Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(getter, namespace, helmDriver, debug)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewInstallerForConfig(cfg), nil
}

func NewInstallerForConfig(cfg *Configuration) *Installer {
	return &Installer{
		cfg: cfg,
	}
}

func (x *Installer) WithOptions(opts InstallOptions) *Installer {
	x.opts = opts
	return x
}

func (x *Installer) WithRegistry(reg *repo.Registry) *Installer {
	x.reg = reg
	return x
}

func (x *Installer) Run() (*release.Release, *flowapi.FlowState, error) {
	if x.opts.Version == "" && x.opts.Devel {
		debug("setting version to >0.0.0-0")
		x.opts.Version = ">0.0.0-0"
	}

	if x.reg == nil {
		return nil, nil, errors.New("x.reg is not set")
	}

	chrt, err := x.reg.GetChart(x.opts.ChartURL, x.opts.ChartName, x.opts.Version)
	if err != nil {
		return nil, nil, err
	}
	// TODO(tamal): Use constant
	setAnnotations(chrt.Chart, "app.kubernetes.io/part-of", x.opts.PartOf)

	cmd := ha.NewInstall(&x.cfg.Configuration)
	var extraAPIs []string

	cmd.DryRun = x.opts.DryRun
	cmd.ReleaseName = x.opts.ReleaseName
	cmd.Namespace = x.opts.Namespace
	cmd.Replace = x.opts.Replace // Skip the name check
	cmd.ClientOnly = false
	cmd.APIVersions = chartutil.VersionSet(extraAPIs)
	cmd.Version = x.opts.Version
	cmd.DisableHooks = x.opts.DisableHooks
	cmd.Wait = x.opts.Wait
	cmd.Timeout = x.opts.Timeout
	cmd.Description = x.opts.Description
	cmd.Atomic = x.opts.Atomic
	cmd.SkipCRDs = x.opts.SkipCRDs
	cmd.SubNotes = x.opts.SubNotes
	cmd.DisableOpenAPIValidation = x.opts.DisableOpenAPIValidation
	cmd.IncludeCRDs = x.opts.IncludeCRDs

	validInstallableChart, err := libchart.IsChartInstallable(chrt.Chart)
	if !validInstallableChart {
		return nil, nil, err
	}

	if chrt.Metadata.Deprecated {
		_, err = fmt.Println("# WARNING: This chart is deprecated")
		if err != nil {
			return nil, nil, err
		}
	}

	if req := chrt.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := ha.CheckDependencies(chrt.Chart, req); err != nil {
			return nil, nil, err
		}
	}

	vals, err := x.opts.Values.MergeValues(chrt.Chart)
	if err != nil {
		return nil, nil, err
	}
	if data, ok := chrt.Chart.Metadata.Annotations["meta.x-helm.dev/editor"]; ok && data != "" {
		var gvr metav1.GroupVersionResource
		if err := json.Unmarshal([]byte(data), &gvr); err != nil {
			return nil, nil, fmt.Errorf("failed to parse %s annotation %s", "meta.x-helm.dev/editor", data)
		}
		rls := types.NamespacedName{
			Namespace: x.opts.Namespace,
			Name:      x.opts.ReleaseName,
		}
		if err := RefillMetadata(hub.NewRegistryOfKnownResources(), chrt.Chart.Values, vals, gvr, rls); err != nil {
			return nil, nil, err
		}
	}
	// chartutil.CoalesceValues(chrt, chrtVals) will use vals to render templates
	chrt.Chart.Values = map[string]interface{}{}

	rls, err := cmd.Run(chrt.Chart, vals)
	if err != nil {
		return nil, nil, err
	}
	caps, _ := x.cfg.GetCapabilities()
	return rls, &flowapi.FlowState{
		ReleaseName:  rls.Name,
		Chrt:         rls.Chart,
		Values:       rls.Config,
		IsUpgrade:    false,
		Capabilities: caps,
		// Engine:       engine.EngineInstance{},
		// InitEngine:   sync.Once{},
	}, nil
}

func (x *Installer) Do() error {
	var err error
	x.result, _, err = x.Run()
	return err
}

func (x *Installer) Result() *release.Release {
	return x.result
}
