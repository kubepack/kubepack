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
	"kubepack.dev/lib-helm/pkg/engine"
	"kubepack.dev/lib-helm/pkg/repo"
	"kubepack.dev/lib-helm/pkg/values"
)

type UpgradeOptions struct {
	ChartURL      string         `json:"chartURL"`
	ChartName     string         `json:"chartName"`
	Version       string         `json:"version"`
	Values        values.Options `json:",inline,omitempty"`
	Install       bool           `json:"install"`
	Devel         bool           `json:"devel"`
	Namespace     string         `json:"namespace"`
	Timeout       time.Duration  `json:"timeout"`
	Wait          bool           `json:"wait"`
	DisableHooks  bool           `json:"disableHooks"`
	DryRun        bool           `json:"dryRun"`
	Force         bool           `json:"force"`
	ResetValues   bool           `json:"resetValues"`
	ReuseValues   bool           `json:"reuseValues"`
	Recreate      bool           `json:"recreate"`
	MaxHistory    int            `json:"maxHistory"`
	Atomic        bool           `json:"atomic"`
	CleanupOnFail bool           `json:"cleanupOnFail"`
	PartOf        string         `json:"partOf"`
}

type Upgrader struct {
	cfg *Configuration

	opts        UpgradeOptions
	reg         *repo.Registry
	releaseName string
	result      *release.Release
}

func NewUpgrader(getter genericclioptions.RESTClientGetter, namespace string, helmDriver string, log ...ha.DebugLog) (*Upgrader, error) {
	cfg := new(Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(getter, namespace, helmDriver, log...)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewUpgraderForConfig(cfg), nil
}

func NewUpgraderForConfig(cfg *Configuration) *Upgrader {
	return &Upgrader{
		cfg: cfg,
	}
}

func (x *Upgrader) WithOptions(opts UpgradeOptions) *Upgrader {
	x.opts = opts
	return x
}

func (x *Upgrader) WithRegistry(reg *repo.Registry) *Upgrader {
	x.reg = reg
	return x
}

func (x *Upgrader) WithReleaseName(name string) *Upgrader {
	x.releaseName = name
	return x
}

func (x *Upgrader) Run() (*release.Release, *engine.State, error) {
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

	cmd := ha.NewUpgrade(&x.cfg.Configuration)
	cmd.Install = x.opts.Install
	cmd.Devel = x.opts.Devel
	cmd.Namespace = x.opts.Namespace
	cmd.Timeout = x.opts.Timeout
	cmd.Wait = x.opts.Wait
	cmd.DisableHooks = x.opts.DisableHooks
	cmd.DryRun = x.opts.DryRun
	cmd.Force = x.opts.Force
	cmd.ResetValues = x.opts.ResetValues
	cmd.ReuseValues = x.opts.ReuseValues
	cmd.Recreate = x.opts.Recreate
	cmd.MaxHistory = x.opts.MaxHistory
	cmd.Atomic = x.opts.Atomic
	cmd.CleanupOnFail = x.opts.CleanupOnFail

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
			Name:      x.releaseName,
		}
		if err := RefillMetadata(hub.NewRegistryOfKnownResources(), chrt.Chart.Values, vals, gvr, rls); err != nil {
			return nil, nil, err
		}
	}
	// chartutil.CoalesceValues(chrt, chrtVals) will use vals to render templates
	chrt.Chart.Values = map[string]interface{}{}

	rls, err := cmd.Run(x.releaseName, chrt.Chart, vals)
	if err != nil {
		return nil, nil, err
	}
	caps, _ := x.cfg.GetCapabilities()
	return rls, &engine.State{
		ReleaseName:  rls.Name,
		Namespace:    x.opts.Namespace,
		Chrt:         rls.Chart,
		Values:       rls.Config,
		IsUpgrade:    true,
		Capabilities: caps,
	}, nil
}

func (x *Upgrader) Do() error {
	var err error
	x.result, _, err = x.Run()
	return err
}

func (x *Upgrader) Result() *release.Release {
	return x.result
}
