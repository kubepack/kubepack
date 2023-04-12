package v1alpha1

import (
	kmapi "kmodules.xyz/client-go/api/v1"
)

func (ref *ChartSourceRef) SetDefaults() *ChartSourceRef {
	if ref.SourceRef.APIGroup == "" {
		ref.SourceRef.APIGroup = "source.toolkit.fluxcd.io"
	}
	if ref.SourceRef.Kind == "" {
		ref.SourceRef.Kind = "HelmRepository"
	} else if ref.SourceRef.Kind == "Legacy" || ref.SourceRef.Kind == "Local" || ref.SourceRef.Kind == "Embed" {
		ref.SourceRef.APIGroup = "charts.x-helm.dev"
	}
	return ref
}

func (ref *ChartSourceFlatRef) FromAPIObject(obj ChartSourceRef) *ChartSourceFlatRef {
	obj.SetDefaults()

	ref.Name = obj.Name
	ref.Version = obj.Version
	ref.SourceAPIGroup = obj.SourceRef.APIGroup
	ref.SourceKind = obj.SourceRef.Kind
	ref.SourceNamespace = obj.SourceRef.Namespace
	ref.SourceName = obj.SourceRef.Name
	return ref
}

func (ref *ChartSourceFlatRef) ToAPIObject() ChartSourceRef {
	obj := ChartSourceRef{
		Name:    ref.Name,
		Version: ref.Version,
		SourceRef: kmapi.TypedObjectReference{
			APIGroup:  ref.SourceAPIGroup,
			Kind:      ref.SourceKind,
			Namespace: ref.SourceNamespace,
			Name:      ref.Name,
		},
	}
	obj.SetDefaults()
	return obj
}
