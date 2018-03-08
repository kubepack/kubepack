package v1alpha1

import (
	"github.com/kubepack/pack-server/apis/manifest"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Dependency) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ResourceTypeDependency + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{"app": "kubepack"},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   manifest.GroupName,
			Version: SchemeGroupVersion.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Singular:   ResourceNameDependency,
				Plural:     ResourceTypeDependency,
				Kind:       ResourceKindDependency,
				ShortNames: []string{"mfs"},
			},
		},
	}
}

func (c Release) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ResourceTypeRelease + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{"app": "kubepack"},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   manifest.GroupName,
			Version: SchemeGroupVersion.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Singular:   ResourceNameRelease,
				Plural:     ResourceTypeRelease,
				Kind:       ResourceKindRelease,
				ShortNames: []string{"rl"},
			},
		},
	}
}
