package repo

import (
	"fmt"
	"io/fs"

	fluxsrc "github.com/fluxcd/source-controller/api/v1"
	"kubepack.dev/lib-helm/pkg/chart/loader"
	releasesapi "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

type IRegistry interface {
	GetChart(srcRef releasesapi.ChartSourceRef) (*ChartExtended, error)
	GetHelmRepository(srcRef releasesapi.ChartSourceRef) (*fluxsrc.HelmRepository, error)
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

func (r EmbeddedRegistry) GetHelmRepository(obj releasesapi.ChartSourceRef) (*fluxsrc.HelmRepository, error) {
	return nil, fmt.Errorf("no helmrepository exists for EmbeddedRegistry")
}
