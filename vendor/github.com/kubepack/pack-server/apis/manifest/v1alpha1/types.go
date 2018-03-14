package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindDependency = "Dependency"
	ResourceNameDependency = "dependency"
	ResourceTypeDependency = "dependencies"
)

const (
	DependencyFile      = "dependency-list.yaml"
	KubepackOpenapiPath = ".kubepack"
	ManifestDirectory   = "manifests"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Dependency struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Package string `json:"package"`
	Version string `json:"version,omitempty"`
	Branch  string `json:"branch,omitempty"`
	Folder  string `json:"folder,omitempty"`
	Repo    string `json:"repo,omitempty"`
	Fork    string `json:"fork,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DependencyList is a list of Dependency objects.
type DependencyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Patches         []string `json:"patches,omitempty"`

	Items []Dependency `json:"items"`
}

const (
	ResourceKindRelease = "Release"
	ResourceNameRelease = "release"
	ResourceTypeRelease = "releases"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ReleaseSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ReleaseStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type ReleaseSpec struct {
}

type ReleaseStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReleaseList is a list of Release objects.
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Release `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ComponentConfig struct {
	metav1.TypeMeta `json:",inline"`

	ClusterName          string `json:"clusterName,omitempty"`
	EnableEventForwarder bool   `json:"enableEventForwarder"`
	ExtractDockerLabel   bool   `json:"extractDockerLabel"`
}
