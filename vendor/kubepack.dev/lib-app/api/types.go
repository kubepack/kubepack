package api

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	metaapi "kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
)

type Metadata struct {
	Resource metaapi.ResourceID `json:"resource,omitempty"`
	Release  ReleaseMetadata    `json:"release,omitempty"`
}

type ReleaseMetadata struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type OptionsSpec struct {
	Metadata `json:"metadata,omitempty"`
}

type ChartOrder struct {
	v1alpha1.ChartRepoRef `json:",inline"`

	ReleaseName string                 `json:"releaseName,omitempty"`
	Namespace   string                 `json:"namespace,omitempty"`
	Values      map[string]interface{} `json:"values,omitempty"`
}

type EditorParameters struct {
	ValuesFile string `json:"valuesFile,omitempty"`
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"valuesPatch,omitempty"`
}

type EditResourceOrder struct {
	Group    string `json:"group,omitempty"`
	Version  string `json:"version,omitempty"`
	Resource string `json:"resource,omitempty"`

	ReleaseName string `json:"releaseName,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Values      string `json:"values,omitempty"`
}

type BucketFile struct {
	// URL of the file in bucket
	URL string `json:"url"`
	// Bucket key for this file
	Key      string `json:"key"`
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
}

type BucketJsonFile struct {
	// URL of the file in bucket
	URL string `json:"url,omitempty"`
	// Bucket key for this file
	Key      string                     `json:"key,omitempty"`
	Filename string                     `json:"filename,omitempty"`
	Data     *unstructured.Unstructured `json:"data,omitempty"`
}

type BucketObject struct {
	// URL of the file in bucket
	URL string `json:"url"`
	// Bucket key for this file
	Key string `json:"key"`
}

type ChartTemplate struct {
	v1alpha1.ChartRef `json:",inline"`
	Version           string                       `json:"version,omitempty"`
	ReleaseName       string                       `json:"releaseName,omitempty"`
	Namespace         string                       `json:"namespace,omitempty"`
	CRDs              []BucketJsonFile             `json:"crds,omitempty"`
	Manifest          *BucketObject                `json:"manifest,omitempty"`
	Resources         []*unstructured.Unstructured `json:"resources,omitempty"`
}

type BucketFileOutput struct {
	// URL of the file in bucket
	URL string `json:"url,omitempty"`
	// Bucket key for this file
	Key      string `json:"key,omitempty"`
	Filename string `json:"filename,omitempty"`
	Data     string `json:"data,omitempty"`
}

type ChartTemplateOutput struct {
	v1alpha1.ChartRef `json:",inline"`
	Version           string             `json:"version,omitempty"`
	ReleaseName       string             `json:"releaseName,omitempty"`
	Namespace         string             `json:"namespace,omitempty"`
	CRDs              []BucketFileOutput `json:"crds,omitempty"`
	Manifest          *BucketObject      `json:"manifest,omitempty"`
	Resources         []string           `json:"resources,omitempty"`
}

type EditorTemplate struct {
	Manifest  []byte                       `json:"manifest,omitempty"`
	Values    map[string]interface{}       `json:"values,omitempty"`
	Resources []*unstructured.Unstructured `json:"resources,omitempty"`
}

type ResourceOutput struct {
	CRDs      []string `json:"crds,omitempty"`
	Resources []string `json:"resources,omitempty"`
}

