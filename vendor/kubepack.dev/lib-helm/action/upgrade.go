package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	libchart "kubepack.dev/lib-helm/chart"
	"kubepack.dev/lib-helm/repo"
	"sigs.k8s.io/yaml"
)

type UpgradeOptions struct {
	ChartURL      string                `json:"chartURL"`
	ChartName     string                `json:"chartName"`
	Version       string                `json:"version"`
	ValuesFile    string                `json:"valuesFile"`
	ValuesPatch   *runtime.RawExtension `json:"valuesPatch"`
	Install       bool                  `json:"install"`
	Devel         bool                  `json:"devel"`
	Namespace     string                `json:"namespace"`
	Timeout       time.Duration         `json:"timeout"`
	Wait          bool                  `json:"wait"`
	DisableHooks  bool                  `json:"disableHooks"`
	DryRun        bool                  `json:"dryRun"`
	Force         bool                  `json:"force"`
	ResetValues   bool                  `json:"resetValues"`
	ReuseValues   bool                  `json:"reuseValues"`
	Recreate      bool                  `json:"recreate"`
	MaxHistory    int                   `json:"maxHistory"`
	Atomic        bool                  `json:"atomic"`
	CleanupOnFail bool                  `json:"cleanupOnFail"`
}

type Upgrader struct {
	cfg *action.Configuration

	opts        UpgradeOptions
	reg         *repo.Registry
	releaseName string
	result      *release.Release
}

func NewUpgrader(getter genericclioptions.RESTClientGetter, namespace string, helmDriver string) (*Upgrader, error) {
	cfg := new(action.Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(getter, namespace, helmDriver, debug)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewUpgraderForConfig(cfg), nil
}

func NewUpgraderForConfig(cfg *action.Configuration) *Upgrader {
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

func (x *Upgrader) Run() (*release.Release, error) {
	if x.opts.Version == "" && x.opts.Devel {
		debug("setting version to >0.0.0-0")
		x.opts.Version = ">0.0.0-0"
	}

	if x.reg == nil {
		return nil, errors.New("x.reg is not set")
	}

	chrt, err := x.reg.GetChart(x.opts.ChartURL, x.opts.ChartName, x.opts.Version)
	if err != nil {
		return nil, err
	}

	cmd := action.NewUpgrade(x.cfg)
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
		return nil, err
	}

	if chrt.Metadata.Deprecated {
		_, err = fmt.Println("# WARNING: This chart is deprecated")
		if err != nil {
			return nil, err
		}
	}

	if req := chrt.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chrt.Chart, req); err != nil {
			return nil, err
		}
	}

	vals := chrt.Values
	if x.opts.ValuesPatch != nil {
		var values []byte
		if x.opts.ValuesFile == "" {
			values, err = json.Marshal(vals)
			if err != nil {
				return nil, err
			}
		} else {
			for _, f := range chrt.Raw {
				if f.Name == x.opts.ValuesFile {
					if err := yaml.Unmarshal(f.Data, &vals); err != nil {
						return nil, fmt.Errorf("cannot load %s. Reason: %v", f.Name, err.Error())
					}
					values, err = yaml.YAMLToJSON(f.Data)
					if err != nil {
						return nil, err
					}
					break
				}
			}
		}

		patchData, err := json.Marshal(x.opts.ValuesPatch)
		if err != nil {
			return nil, err
		}
		patch, err := jsonpatch.DecodePatch(patchData)
		if err != nil {
			return nil, err
		}
		modifiedValues, err := patch.Apply(values)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(modifiedValues, &vals)
		if err != nil {
			return nil, err
		}
	}

	return cmd.Run(x.releaseName, chrt.Chart, vals)
}

func (x *Upgrader) Do() error {
	var err error
	x.result, err = x.Run()
	return err
}

func (x *Upgrader) Result() *release.Release {
	return x.result
}
