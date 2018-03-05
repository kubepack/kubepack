package v1alpha1

import (
	"github.com/kubepack/packserver/apis/apps"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Manifest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ResourceTypeManifest + "." + SchemeGroupVersion.Group,
			Labels: map[string]string{"app": "kubepack"},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   apps.GroupName,
			Version: SchemeGroupVersion.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Singular:   ResourceNameManifest,
				Plural:     ResourceTypeManifest,
				Kind:       ResourceKindManifest,
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
			Group:   apps.GroupName,
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
