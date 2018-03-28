package v1alpha1

import (
	"github.com/kubepack/pack-server/apis/manifest"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Dependency) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ResourcePluralDependency + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{"app": "kubepack"},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   manifest.GroupName,
			Version: SchemeGroupVersion.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     ResourcePluralDependency,
				Singular:   ResourceSingularDependency,
				Kind:       ResourceKindDependency,
				ShortNames: []string{"dep"},
			},
		},
	}
}

func (c Release) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ResourcePluralRelease + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{"app": "kubepack"},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   manifest.GroupName,
			Version: SchemeGroupVersion.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     ResourcePluralRelease,
				Singular:   ResourceSingularRelease,
				Kind:       ResourceKindRelease,
				ShortNames: []string{"rl"},
			},
		},
	}
}

const (
	ResourceKindManifest     = "Manifest"
	ResourcePluralManifest   = "manifests"
	ResourceSingularManifest = "manifest"
)

func ManifestResourceDefinition() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ResourcePluralManifest + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{"app": "kubepack"},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   manifest.GroupName,
			Version: SchemeGroupVersion.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     ResourcePluralManifest,
				Singular:   ResourceSingularManifest,
				Kind:       ResourceKindManifest,
				ShortNames: []string{"mfs"},
			},
		},
	}
}
