// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package statusreaders

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/engine"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"
)

// NewDefaultStatusReader returns a DelegatingStatusReader that wraps a list of
// statusreaders to cover all built-in Kubernetes resources and other CRDs that
// follow known status conventions.
func NewDefaultStatusReader(mapper meta.RESTMapper) engine.StatusReader {
	defaultStatusReader := NewGenericStatusReader(mapper, status.Compute)

	replicaSetStatusReader := NewReplicaSetStatusReader(mapper, defaultStatusReader)
	deploymentStatusReader := NewDeploymentResourceReader(mapper, replicaSetStatusReader)
	statefulSetStatusReader := NewStatefulSetResourceReader(mapper, defaultStatusReader)

	return &DelegatingStatusReader{
		StatusReaders: []engine.StatusReader{
			deploymentStatusReader,
			statefulSetStatusReader,
			replicaSetStatusReader,
			defaultStatusReader,
		},
	}
}

type DelegatingStatusReader struct {
	StatusReaders []engine.StatusReader
}

func (dsr *DelegatingStatusReader) Supports(gk schema.GroupKind) bool {
	for _, sr := range dsr.StatusReaders {
		if sr.Supports(gk) {
			return true
		}
	}
	return false
}

func (dsr *DelegatingStatusReader) ReadStatus(
	ctx context.Context,
	reader engine.ClusterReader,
	id object.ObjMetadata,
) (*event.ResourceStatus, error) {
	gk := id.GroupKind
	for _, sr := range dsr.StatusReaders {
		if sr.Supports(gk) {
			return sr.ReadStatus(ctx, reader, id)
		}
	}
	return nil, fmt.Errorf("no status reader supports this resource: %v", gk)
}

func (dsr *DelegatingStatusReader) ReadStatusForObject(
	ctx context.Context,
	reader engine.ClusterReader,
	obj *unstructured.Unstructured,
) (*event.ResourceStatus, error) {
	gk := obj.GroupVersionKind().GroupKind()
	for _, sr := range dsr.StatusReaders {
		if sr.Supports(gk) {
			return sr.ReadStatusForObject(ctx, reader, obj)
		}
	}
	return nil, fmt.Errorf("no status reader supports this resource: %v", gk)
}
