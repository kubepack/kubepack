package action

import (
	"time"

	ha "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type UninstallOptions struct {
	DisableHooks bool          `json:"disableHooks"`
	DryRun       bool          `json:"dryRun"`
	KeepHistory  bool          `json:"keepHistory"`
	Timeout      time.Duration `json:"timeout"`
}

type Uninstaller struct {
	cfg *Configuration

	opts        UninstallOptions
	releaseName string
	result      *release.UninstallReleaseResponse
}

func NewUninstaller(getter genericclioptions.RESTClientGetter, namespace string, helmDriver string, log ...ha.DebugLog) (*Uninstaller, error) {
	cfg := new(Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(getter, namespace, helmDriver, log...)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewUninstallerForConfig(cfg), nil
}

func NewUninstallerForConfig(cfg *Configuration) *Uninstaller {
	return &Uninstaller{
		cfg: cfg,
	}
}

func (x *Uninstaller) WithOptions(opts UninstallOptions) *Uninstaller {
	x.opts = opts
	return x
}

func (x *Uninstaller) WithReleaseName(name string) *Uninstaller {
	x.releaseName = name
	return x
}

func (x *Uninstaller) Run() (*release.UninstallReleaseResponse, error) {
	cmd := ha.NewUninstall(&x.cfg.Configuration)
	cmd.DisableHooks = x.opts.DisableHooks
	cmd.DryRun = x.opts.DryRun
	cmd.KeepHistory = x.opts.KeepHistory
	cmd.Timeout = x.opts.Timeout

	return cmd.Run(x.releaseName)
}

func (x *Uninstaller) Do() error {
	var err error
	x.result, err = x.Run()
	return err
}

func (x *Uninstaller) Result() *release.UninstallReleaseResponse {
	return x.result
}
