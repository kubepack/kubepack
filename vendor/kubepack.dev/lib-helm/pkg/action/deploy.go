package action

import (
	"time"

	ha "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"kubepack.dev/lib-helm/pkg/repo"
	"kubepack.dev/lib-helm/pkg/values"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

type DeployOptions struct {
	releasesapi.ChartSourceFlatRef `json:",inline,omitempty"`
	values.Options                 `json:",inline,omitempty"`
	ClientOnly                     bool          `json:"clientOnly"`
	DryRun                         bool          `json:"dryRun"`
	DisableHooks                   bool          `json:"disableHooks"`
	Replace                        bool          `json:"replace"`
	Wait                           bool          `json:"wait"`
	Devel                          bool          `json:"devel"`
	Timeout                        time.Duration `json:"timeout"`
	Namespace                      string        `json:"namespace"`
	ReleaseName                    string        `json:"releaseName"`
	Description                    string        `json:"description"`
	Atomic                         bool          `json:"atomic"`
	SkipCRDs                       bool          `json:"skipCRDs"`
	SubNotes                       bool          `json:"subNotes"`
	DisableOpenAPIValidation       bool          `json:"disableOpenAPIValidation"`
	IncludeCRDs                    bool          `json:"includeCRDs"`
	PartOf                         string        `json:"partOf"`
	CreateNamespace                bool          `json:"createNamespace"`
	Force                          bool          `json:"force"`
	Recreate                       bool          `json:"recreate"`
	ResetValues                    bool          `json:"resetValues"`
	ReuseValues                    bool          `json:"reuseValues"`
	CleanupOnFail                  bool          `json:"cleanupOnFail"`
}

type Deployer struct {
	cfg *Configuration

	opts   DeployOptions
	reg    repo.IRegistry
	result *release.Release
}

func NewDeployer(getter genericclioptions.RESTClientGetter, namespace string, helmDriver string, log ...ha.DebugLog) (*Deployer, error) {
	cfg := new(Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(getter, namespace, helmDriver, log...)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewDeployerForConfig(cfg), nil
}

func NewDeployerForConfig(cfg *Configuration) *Deployer {
	return &Deployer{
		cfg: cfg,
	}
}

func (x *Deployer) WithOptions(opts DeployOptions) *Deployer {
	x.opts = opts
	return x
}

func (x *Deployer) WithRegistry(reg repo.IRegistry) *Deployer {
	x.reg = reg
	return x
}

func (x *Deployer) Run() (*release.Release, error) {
	// If a release does not exist, install it.
	histClient := ha.NewHistory(&x.cfg.Configuration)
	histClient.Max = 1
	if _, err := histClient.Run(x.opts.ReleaseName); err == driver.ErrReleaseNotFound {
		i := NewInstallerForConfig(x.cfg)
		i.WithRegistry(x.reg).
			WithOptions(InstallOptions{
				ChartSourceFlatRef:       x.opts.ChartSourceFlatRef,
				Options:                  x.opts.Options,
				ClientOnly:               x.opts.ClientOnly,
				DryRun:                   x.opts.DryRun,
				DisableHooks:             x.opts.DisableHooks,
				Replace:                  false,
				Wait:                     x.opts.Wait,
				Devel:                    x.opts.Devel,
				Timeout:                  x.opts.Timeout,
				Namespace:                x.opts.Namespace,
				ReleaseName:              x.opts.ReleaseName,
				Description:              x.opts.Description,
				Atomic:                   x.opts.Atomic,
				SkipCRDs:                 x.opts.SkipCRDs,
				SubNotes:                 x.opts.SubNotes,
				DisableOpenAPIValidation: x.opts.DisableOpenAPIValidation,
				IncludeCRDs:              x.opts.IncludeCRDs,
				PartOf:                   x.opts.PartOf,
				CreateNamespace:          x.opts.CreateNamespace,
			})
		return i.Run()
	} else if err != nil {
		return nil, err
	}

	i := NewUpgraderForConfig(x.cfg)
	i.WithRegistry(x.reg).
		WithReleaseName(x.opts.ReleaseName).
		WithOptions(UpgradeOptions{
			ChartSourceFlatRef: x.opts.ChartSourceFlatRef,
			Options:            x.opts.Options,
			Install:            false,
			Devel:              x.opts.Devel,
			Namespace:          x.opts.Namespace,
			Timeout:            x.opts.Timeout,
			Wait:               x.opts.Wait,
			DisableHooks:       x.opts.DisableHooks,
			DryRun:             x.opts.DryRun,
			Force:              x.opts.Force,
			ResetValues:        x.opts.ResetValues,
			ReuseValues:        x.opts.ReuseValues,
			Recreate:           x.opts.Recreate,
			MaxHistory:         0,
			Atomic:             x.opts.Atomic,
			CleanupOnFail:      x.opts.CleanupOnFail,
			PartOf:             x.opts.PartOf,
		})
	return i.Run()
}

func (x *Deployer) Do() error {
	var err error
	x.result, err = x.Run()
	return err
}

func (x *Deployer) Result() *release.Release {
	return x.result
}
