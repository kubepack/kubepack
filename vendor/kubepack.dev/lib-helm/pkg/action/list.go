package action

import (
	ha "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ListOptions struct {
	All          bool      `json:"all"`
	Namespace    string    `json:"namespace"`
	Sort         ha.Sorter `json:"sort"`
	ByDate       bool      `json:"byDate"`
	SortReverse  bool      `json:"sortReverse"`
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
	Filter       string    `json:"filter"`
	Short        bool      `json:"short"`
	Uninstalled  bool      `json:"uninstalled"`
	Superseded   bool      `json:"superseded"`
	Uninstalling bool      `json:"uninstalling"`
	Deployed     bool      `json:"deployed"`
	Failed       bool      `json:"failed"`
	Pending      bool      `json:"pending"`
}

type Lister struct {
	cfg *Configuration

	opts   ListOptions
	result []*release.Release
}

func NewLister(getter genericclioptions.RESTClientGetter, namespace string, helmDriver string, log ...ha.DebugLog) (*Lister, error) {
	cfg := new(Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(getter, namespace, helmDriver, log...)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewListerForConfig(cfg), nil
}

func NewListerForConfig(cfg *Configuration) *Lister {
	return &Lister{
		cfg: cfg,
	}
}

func (x *Lister) WithOptions(opts ListOptions) *Lister {
	x.opts = opts
	return x
}

func (x *Lister) Run() ([]*release.Release, error) {
	cmd := ha.NewList(&x.cfg.Configuration)
	cmd.All = x.opts.All
	cmd.AllNamespaces = x.opts.Namespace == ""
	cmd.Sort = x.opts.Sort
	cmd.ByDate = x.opts.ByDate
	cmd.SortReverse = x.opts.SortReverse
	cmd.Limit = x.opts.Limit
	cmd.Offset = x.opts.Offset
	cmd.Filter = x.opts.Filter
	cmd.Short = x.opts.Short
	cmd.Uninstalled = x.opts.Uninstalled
	cmd.Superseded = x.opts.Superseded
	cmd.Uninstalling = x.opts.Uninstalling
	cmd.Deployed = x.opts.Deployed
	cmd.Failed = x.opts.Failed
	cmd.Pending = x.opts.Pending

	cmd.SetStateMask()
	return cmd.Run()
}

func (x *Lister) Do() error {
	var err error
	x.result, err = x.Run()
	return err
}

func (x *Lister) Result() []*release.Release {
	return x.result
}
