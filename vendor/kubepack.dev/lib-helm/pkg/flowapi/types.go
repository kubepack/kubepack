package flowapi

import (
	"sync"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	rsapi "kmodules.xyz/resource-metadata/apis/meta/v1alpha1"
)

type FlowState struct {
	ReleaseName  string
	Chrt         *chart.Chart
	Values       chartutil.Values // final values used for rendering
	IsUpgrade    bool
	Capabilities *chartutil.Capabilities

	Engine     *engine.EngineInstance
	InitEngine sync.Once
}

type Flow struct {
	Name     string            `json:"name"` // should be metadata.name
	Actions  []Action          `json:"actions"`
	EdgeList []rsapi.NamedEdge `json:"edge_list"`
}

// Check array, map, etc
// can this be always string like in --set keys?
// Keep is such that we can always generate helm equivalent command
type KV struct {
	Key string
	// string, nil, null
	Type string `json:"type"`
	// format is an optional OpenAPI type definition for this column. The 'name' format is applied
	// to the primary identifier column to assist in clients identifying column is the resource name.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	// +optional
	// Format string `json:"format,omitempty"`

	// PathTemplate is a Go text template that will be evaluated to determine cell value.
	// Users can use JSONPath expression to extract nested fields and apply template functions from Masterminds/sprig library.
	// The template function for JSON path is called `jp`.
	// Example: {{ jp "{.a.b}" . }} or {{ jp "{.a.b}" true }}, if json output is desired from JSONPath parser
	// +optional
	PathTemplate string `json:"pathTemplate,omitempty"`
	//
	//
	// Directly use path from object
	Path string `json:"path"`

	// json patch operation
	// See also: http://jsonpatch.com/
	// Op string `json:"op"`
}

type LoadValue struct {
	From   ObjectLocator `json:"from"`
	Values []KV          `json:"values"`
}

type ObjectLocator struct {
	// Use the values from that release == action to render templates
	UseRelease string    `json:"use_release"`
	Src        ObjectRef `json:"src"`
	Paths      []string  `json:"paths"` // sequence of DirectedEdge names
}

type ObjectRef struct {
	Target       metav1.TypeMeta       `json:"target"`
	Selector     *metav1.LabelSelector `json:"selector,omitempty"`
	Name         string                `json:"name,omitempty"`
	NameTemplate string                `json:"nameTemplate,omitempty"`
	// Namespace always same as Workflow
}

type Action struct {
	// Also the action name
	ReleaseName string `json:"releaseName" protobuf:"bytes,3,opt,name=releaseName"`

	rsapi.ChartRepoRef `json:",inline" protobuf:"bytes,1,opt,name=chartRef"`

	// Namespace   string `json:"namespace" protobuf:"bytes,4,opt,name=namespace"`

	ValuesFile string `json:"valuesFile,omitempty" protobuf:"bytes,6,opt,name=valuesFile"`
	// RFC 6902 compatible json patch. ref: http://jsonpatch.com
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ValuesPatch *runtime.RawExtension `json:"valuesPatch,omitempty" protobuf:"bytes,7,opt,name=valuesPatch"`

	ValueOverrides []LoadValue `json:"overrideValues"`

	// https://github.com/tamalsaha/kstatus-demo
	ReadinessCriteria ReadinessCriteria `json:"readiness_criteria"`

	Prerequisites Prerequisites `json:"prerequisites"`
}

type Prerequisites struct {
	RequiredResources []metav1.GroupVersionResource `json:"required_resources"`
}

type ReadinessCriteria struct {
	Timeout metav1.Duration `json:"timeout"`

	// List objects for which to wait to reconcile using kstatus == Current
	// Same as helm --wait
	WaitForReconciled bool `json:"wait_for_reconciled"`

	ResourcesExist []metav1.GroupVersionResource `json:"required_resources"`
	WaitFors       []WaitFlags                   `json:"waitFors,omitempty" protobuf:"bytes,9,rep,name=waitFors"`
}

type ChartRef struct {
	URL  string `json:"url" protobuf:"bytes,1,opt,name=url"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

type ResourceDefinitions struct {
	Owned    []metav1.GroupVersionResource `json:"owned" protobuf:"bytes,1,rep,name=owned"`
	Required []metav1.GroupVersionResource `json:"required" protobuf:"bytes,2,rep,name=required"`
}

// wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available]

type WaitFlags struct {
	Resource     metav1.GroupResource  `json:"resource" protobuf:"bytes,1,opt,name=resource"`
	Labels       *metav1.LabelSelector `json:"labels" protobuf:"bytes,2,opt,name=labels"`
	All          bool                  `json:"all" protobuf:"varint,3,opt,name=all"`
	Timeout      metav1.Duration       `json:"timeout" protobuf:"bytes,4,opt,name=timeout"`
	ForCondition string                `json:"for" protobuf:"bytes,5,opt,name=for"`
}
