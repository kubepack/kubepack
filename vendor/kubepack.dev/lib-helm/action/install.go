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

type InstallOptions struct {
	ChartURL     string                `json:"chartURL"`
	ChartName    string                `json:"chartName"`
	Version      string                `json:"version"`
	ValuesFile   string                `json:"valuesFile"`
	ValuesPatch  *runtime.RawExtension `json:"valuesPatch"`
	DryRun       bool                  `json:"dryRun"`
	DisableHooks bool                  `json:"disableHooks"`
	Replace      bool                  `json:"replace"`
	Wait         bool                  `json:"wait"`
	Devel        bool                  `json:"devel"`
	Timeout      time.Duration         `json:"timeout"`
	Namespace    string                `json:"namespace"`
	ReleaseName  string                `json:"releaseName"`
	Atomic       bool                  `json:"atomic"`
	SkipCRDs     bool                  `json:"skipCRDs"`
}

type Installer struct {
	cfg *action.Configuration

	opts   InstallOptions
	reg    *repo.Registry
	result *release.Release
}

func NewInstaller(getter genericclioptions.RESTClientGetter, namespace string, helmDriver string) (*Installer, error) {
	cfg := new(action.Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(getter, namespace, helmDriver, debug)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewInstallerForConfig(cfg), nil
}

func NewInstallerForConfig(cfg *action.Configuration) *Installer {
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

func (x *Installer) Run() (*release.Release, error) {
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

	cmd := action.NewInstall(x.cfg)
	var extraAPIs []string

	cmd.DryRun = x.opts.DryRun
	cmd.ReleaseName = x.opts.ReleaseName
	cmd.Namespace = x.opts.Namespace
	cmd.Replace = x.opts.Replace // Skip the name check
	cmd.ClientOnly = false
	cmd.APIVersions = chartutil.VersionSet(extraAPIs)
	cmd.Version = x.opts.Version
	cmd.DisableHooks = x.opts.DisableHooks
	cmd.Atomic = x.opts.Atomic
	cmd.Wait = x.opts.Wait
	cmd.Timeout = x.opts.Timeout

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

	return cmd.Run(chrt.Chart, vals)
}

func (x *Installer) Do() error {
	var err error
	x.result, err = x.Run()
	return err
}

func (x *Installer) Result() *release.Release {
	return x.result
}
