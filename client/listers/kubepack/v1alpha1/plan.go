/*
Copyright The Kubepack Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PlanLister helps list Plans.
type PlanLister interface {
	// List lists all Plans in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Plan, err error)
	// Get retrieves the Plan from the index for a given name.
	Get(name string) (*v1alpha1.Plan, error)
	PlanListerExpansion
}

// planLister implements the PlanLister interface.
type planLister struct {
	indexer cache.Indexer
}

// NewPlanLister returns a new PlanLister.
func NewPlanLister(indexer cache.Indexer) PlanLister {
	return &planLister{indexer: indexer}
}

// List lists all Plans in the indexer.
func (s *planLister) List(selector labels.Selector) (ret []*v1alpha1.Plan, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Plan))
	})
	return ret, err
}

// Get retrieves the Plan from the index for a given name.
func (s *planLister) Get(name string) (*v1alpha1.Plan, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("plan"), name)
	}
	return obj.(*v1alpha1.Plan), nil
}