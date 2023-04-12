package repo

import (
	"fmt"
	"io/fs"

	"kubepack.dev/lib-helm/pkg/chart/loader"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

type IRegistry interface {
	GetChart(srcRef releasesapi.ChartSourceRef) (*ChartExtended, error)
}

type EmbeddedRegistry struct {
	rootFS fs.FS
}

func NewEmbeddedRegistry(fsys fs.FS) IRegistry {
	return &EmbeddedRegistry{rootFS: fsys}
}

func (r EmbeddedRegistry) GetChart(srcRef releasesapi.ChartSourceRef) (*ChartExtended, error) {
	if srcRef.SourceRef.Kind != releasesapi.SourceKindEmbed {
		return nil, fmt.Errorf("invalid source kind, expected Embed, found: %s", srcRef.SourceRef.Kind)
	}

	name := srcRef.Name
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
