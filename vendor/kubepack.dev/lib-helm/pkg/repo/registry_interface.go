package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kubepack.dev/lib-helm/pkg/chart/loader"
)

type IRegistry interface {
	Register(srcRef kmapi.TypedObjectReference) (string, error) // url, error
	GetChart(repository, chartName, chartVersion string) (*ChartExtended, error)
}

type EmbeddedRegistry struct {
	rootFS fs.FS
}

func NewEmbeddedRegistry(fsys fs.FS) IRegistry {
	return &EmbeddedRegistry{rootFS: fsys}
}

func (r EmbeddedRegistry) Register(_ kmapi.TypedObjectReference) (string, error) {
	return "", errors.New("unsupported method")
}

func (r EmbeddedRegistry) GetChart(repository, chartName, _ string) (*ChartExtended, error) {
	name, embedded := IsEmbedded(repository)
	if !embedded {
		return nil, fmt.Errorf("invalid repository format, expected embed://{chartName}, found: %s", repository)
	}
	if chartName != "" && chartName != name {
		return nil, fmt.Errorf("invalid chartname, expected %s, found: %s", name, chartName)
	}
	if name == "" {
		name = "."
	}
	fsys, err := fs.Sub(r.rootFS, name)
	if err != nil {
		return nil, err
	}

	chrt, err := loader.LoadFS(fsys)
	if err != nil {
		return nil, err
	}

	return &ChartExtended{
		Chart: chrt,
	}, nil
}

func IsEmbedded(repository string) (chartName string, embedded bool) {
	repository = strings.TrimSpace(repository)
	return strings.TrimPrefix(repository, "embed:///"), strings.HasPrefix(repository, "embed:///")
}
