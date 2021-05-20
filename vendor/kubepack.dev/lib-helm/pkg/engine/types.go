package engine

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

type State struct {
	ReleaseName  string
	Namespace    string
	Chrt         *chart.Chart
	Values       chartutil.Values // final values used for rendering
	IsUpgrade    bool
	Capabilities *chartutil.Capabilities

	Engine *engine.EngineInstance
}

func (state State) Init() error {
	if state.Engine != nil {
		return nil
	}

	options := chartutil.ReleaseOptions{
		Name:      state.ReleaseName,
		Namespace: state.Namespace,
		Revision:  1,
		IsInstall: !state.IsUpgrade,
		IsUpgrade: state.IsUpgrade,
	}
	valuesToRender, err := chartutil.ToRenderValues(state.Chrt, state.Values, options, state.Capabilities)
	if err != nil {
		return errors.Wrap(err, "failed to initialize engine")
	}
	state.Engine = new(engine.Engine).NewInstance(state.Chrt, valuesToRender) // reuse engine
	return nil
}
