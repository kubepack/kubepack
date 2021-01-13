package action

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"kubepack.dev/cli/pkg/action"
	"kubepack.dev/cli/pkg/apply"
	"kubepack.dev/lib-helm/repo"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"
)

type ApplyOptions struct {
	*apply.ApplyOptions

	ChartURL  string `json:"chartURL"`
	ChartName string `json:"chartName"`
	Version   string `json:"version"`

	ValuesFile  string                `json:"valuesFile"`
	ValuesPatch *runtime.RawExtension `json:"valuesPatch"`
	// OR
	Values map[string]interface{} `json:"values"`

	CreateNamespace bool
	DryRun          bool
	DisableHooks    bool
	Replace         bool
	Wait            bool
	Timeout         time.Duration
	Description     string

	Devel bool
	// DependencyUpdate         bool
	Namespace   string
	ReleaseName string

	//OutputDir                string
	Atomic                   bool
	SkipCRDs                 bool
	SubNotes                 bool
	DisableOpenAPIValidation bool
	IncludeCRDs              bool
}

func NewApplyOptions() ApplyOptions {
	return ApplyOptions{
		ApplyOptions: apply.NewApplyOptions(genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		}),
	}
}

type Applier struct {
	cfg *action.Configuration

	f      cmdutil.Factory
	opts   ApplyOptions
	reg    *repo.Registry
	result *release.Release
}

func NewApplier(f cmdutil.Factory, namespace string, helmDriver string) (*Applier, error) {
	cfg := new(action.Configuration)
	// TODO: Use secret driver for which namespace?
	err := cfg.Init(f, namespace, helmDriver, debug)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = chartutil.DefaultCapabilities

	return NewApplierForConfig(f, cfg), nil
}

func NewApplierForConfig(f cmdutil.Factory, cfg *action.Configuration) *Applier {
	return &Applier{
		f:   f,
		cfg: cfg,
	}
}

func (x *Applier) WithOptions(opts ApplyOptions) *Applier {
	x.opts = opts
	return x
}

func (x *Applier) WithRegistry(reg *repo.Registry) *Applier {
	x.reg = reg
	return x
}

// args []string, cmd *action.Apply, valueOpts *values.Options, out io.Writer
func (x *Applier) Run() (*release.Release, error) {
	debug("Original chart version: %q", x.opts.Version)
	if x.opts.Version == "" && x.opts.Devel {
		debug("setting version to >0.0.0-0")
		x.opts.Version = ">0.0.0-0"
	}

	if x.reg == nil {
		return nil, errors.New("x.reg is not set")
	}

	err := x.opts.Complete(x.f)
	if err != nil {
		return nil, err
	}

	chrt, err := x.reg.GetChart(x.opts.ChartURL, x.opts.ChartName, x.opts.Version)
	if err != nil {
		return nil, err
	}

	cmd := action.NewApply(x.cfg)

	cmd.ApplyOptions = x.opts.ApplyOptions
	cmd.CreateNamespace = x.opts.CreateNamespace
	cmd.DryRun = x.opts.DryRun
	cmd.DisableHooks = x.opts.DisableHooks
	cmd.Replace = x.opts.Replace
	cmd.Wait = x.opts.Wait
	cmd.Timeout = x.opts.Timeout
	cmd.Description = x.opts.Description
	cmd.Devel = x.opts.Devel
	// cmd.DependencyUpdate         = x.opts.DependencyUpdate
	cmd.Namespace = x.opts.Namespace
	cmd.ReleaseName = x.opts.ReleaseName
	//OutputDir
	cmd.Atomic = x.opts.Atomic
	cmd.SkipCRDs = x.opts.SkipCRDs
	cmd.SubNotes = x.opts.SubNotes
	cmd.DisableOpenAPIValidation = x.opts.DisableOpenAPIValidation
	cmd.IncludeCRDs = x.opts.IncludeCRDs

	validApplyableChart, err := isChartApplyable(chrt.Chart)
	if !validApplyableChart {
		return nil, err
	}

	if chrt.Chart.Metadata.Deprecated {
		_, _ = fmt.Fprintln(x.opts.Out, "WARNING: This chart is deprecated")
	}

	if req := chrt.Chart.Metadata.Dependencies; req != nil {
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
	} else if x.opts.Values != nil {
		vals = x.opts.Values
	}
	// VERY IMP: changes the default rendering behavior of helm charts
	cmd.OverwriteValues = x.opts.Values != nil

	return cmd.Run(chrt.Chart, vals)
}

// isChartApplyable validates if a chart can be applyed
//
// Application chart type is only applyable
func isChartApplyable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not applyable", ch.Metadata.Type)
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	_ = log.Output(2, fmt.Sprintf(format, v...))
}
