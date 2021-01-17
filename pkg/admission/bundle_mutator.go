/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package admission

import (
	"sync"

	api "kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

const (
	mutatorGroup   = "mutators.kubepack.com"
	mutatorVersion = "v1alpha1"
)

type BundleMutator struct {
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &BundleMutator{}

func (v *BundleMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    mutatorGroup,
			Version:  mutatorVersion,
			Resource: "bundlemutators",
		},
		"bundlemutator"
}

func (v *BundleMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.initialized = true
	return nil
}

func (v *BundleMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindBundle {
		status.Allowed = true
		return status
	}

	v.lock.RLock()
	defer v.lock.RUnlock()
	if !v.initialized {
		return hookapi.StatusUninitialized()
	}

	obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
	if err != nil {
		return hookapi.StatusBadRequest(err)
	}
	mod := setBundleDefaultValues(obj.(*api.Bundle).DeepCopy())
	patch, err := meta_util.CreateJSONPatch(req.Object.Raw, mod)
	if err != nil {
		return hookapi.StatusInternalServerError(err)
	}
	status.Patch = patch
	patchType := admission.PatchTypeJSONPatch
	status.PatchType = &patchType

	status.Allowed = true
	return status
}

func setBundleDefaultValues(vs *api.Bundle) *api.Bundle {
	return vs
}
