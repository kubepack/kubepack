/*
Copyright The Helm Authors.

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

package action

import (
	"context"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// lazyClient is a workaround to deal with Kubernetes having an unstable kb API.
// In Kubernetes v1.18 the defaults where removed which broke creating a
// kb without an explicit configuration. ಠ_ಠ
type lazyClient struct {
	// kb caches an initialized kubernetes kb
	initClient sync.Once
	kc         kubernetes.Interface
	kb         client.Client
	clientErr  error

	// clientFn loads a kubernetes client
	kcFn func() (*kubernetes.Clientset, error)
	kbFn func() (client.Client, error)

	// namespace passed to each kb request
	namespace string
}

func (s *lazyClient) init() error {
	s.initClient.Do(func() {
		s.kc, s.clientErr = s.kcFn()
		s.kb, s.clientErr = s.kbFn()
	})
	return s.clientErr
}

// appReleaseClient implements a coreappv1beta1.AppReleaseInterface
type appReleaseClient struct{ *lazyClient }

func (a *appReleaseClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := a.init(); err != nil {
		return err
	}
	obj.SetNamespace(a.namespace)
	return a.kb.Get(ctx, key, obj, opts...)
}

func (a *appReleaseClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if err := a.init(); err != nil {
		return err
	}
	return a.kb.List(ctx, list, append(opts, client.InNamespace(a.namespace))...)
}

func (a *appReleaseClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := a.init(); err != nil {
		return err
	}
	obj.SetNamespace(a.namespace)
	return a.kb.Create(ctx, obj, opts...)
}

func (a *appReleaseClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if err := a.init(); err != nil {
		return err
	}
	obj.SetNamespace(a.namespace)
	return a.kb.Delete(ctx, obj, opts...)
}

func (a *appReleaseClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := a.init(); err != nil {
		return err
	}
	obj.SetNamespace(a.namespace)
	return a.kb.Update(ctx, obj, opts...)
}

func (a *appReleaseClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if err := a.init(); err != nil {
		return err
	}
	obj.SetNamespace(a.namespace)
	return a.kb.Patch(ctx, obj, patch, opts...)
}

func (a *appReleaseClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	if err := a.init(); err != nil {
		return err
	}
	return a.kb.DeleteAllOf(ctx, obj, append(opts, client.InNamespace(a.namespace))...)
}

func (a *appReleaseClient) Status() client.SubResourceWriter {
	if err := a.init(); err != nil {
		panic(err)
	}
	return a.kb.Status()
}

func (a *appReleaseClient) SubResource(subResource string) client.SubResourceClient {
	if err := a.init(); err != nil {
		panic(err)
	}
	return a.kb.SubResource(subResource)
}

func (a *appReleaseClient) Scheme() *runtime.Scheme {
	if err := a.init(); err != nil {
		panic(err)
	}
	return a.kb.Scheme()
}

func (a *appReleaseClient) RESTMapper() meta.RESTMapper {
	if err := a.init(); err != nil {
		panic(err)
	}
	return a.kb.RESTMapper()
}

func (a *appReleaseClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	if err := a.init(); err != nil {
		return schema.GroupVersionKind{}, err
	}
	return a.kb.GroupVersionKindFor(obj)
}

func (a *appReleaseClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	if err := a.init(); err != nil {
		return false, err
	}
	return a.kb.IsObjectNamespaced(obj)
}

var _ client.Client = (*appReleaseClient)(nil)

func newAppReleaseClient(lc *lazyClient) *appReleaseClient {
	return &appReleaseClient{lazyClient: lc}
}

// ------------ COPY from HELM

// secretClient implements a corev1.SecretsInterface
type secretClient struct{ *lazyClient }

var _ corev1.SecretInterface = (*secretClient)(nil)

func newSecretClient(lc *lazyClient) *secretClient {
	return &secretClient{lazyClient: lc}
}

func (s *secretClient) Create(ctx context.Context, secret *v1.Secret, opts metav1.CreateOptions) (result *v1.Secret, err error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	return s.kc.CoreV1().Secrets(s.namespace).Create(ctx, secret, opts)
}

func (s *secretClient) Update(ctx context.Context, secret *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	return s.kc.CoreV1().Secrets(s.namespace).Update(ctx, secret, opts)
}

func (s *secretClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	if err := s.init(); err != nil {
		return err
	}
	return s.kc.CoreV1().Secrets(s.namespace).Delete(ctx, name, opts)
}

func (s *secretClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	if err := s.init(); err != nil {
		return err
	}
	return s.kc.CoreV1().Secrets(s.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (s *secretClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Secret, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	return s.kc.CoreV1().Secrets(s.namespace).Get(ctx, name, opts)
}

func (s *secretClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.SecretList, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	return s.kc.CoreV1().Secrets(s.namespace).List(ctx, opts)
}

func (s *secretClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	return s.kc.CoreV1().Secrets(s.namespace).Watch(ctx, opts)
}

func (s *secretClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*v1.Secret, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	return s.kc.CoreV1().Secrets(s.namespace).Patch(ctx, name, pt, data, opts, subresources...)
}

func (s *secretClient) Apply(ctx context.Context, secretConfiguration *applycorev1.SecretApplyConfiguration, opts metav1.ApplyOptions) (*v1.Secret, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	return s.kc.CoreV1().Secrets(s.namespace).Apply(ctx, secretConfiguration, opts)
}

// configMapClient implements a corev1.ConfigMapInterface
type configMapClient struct{ *lazyClient }

var _ corev1.ConfigMapInterface = (*configMapClient)(nil)

func newConfigMapClient(lc *lazyClient) *configMapClient {
	return &configMapClient{lazyClient: lc}
}

func (c *configMapClient) Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).Create(ctx, configMap, opts)
}

func (c *configMapClient) Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).Update(ctx, configMap, opts)
}

func (c *configMapClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	if err := c.init(); err != nil {
		return err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).Delete(ctx, name, opts)
}

func (c *configMapClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	if err := c.init(); err != nil {
		return err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (c *configMapClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).Get(ctx, name, opts)
}

func (c *configMapClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.ConfigMapList, error) {
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).List(ctx, opts)
}

func (c *configMapClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).Watch(ctx, opts)
}

func (c *configMapClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*v1.ConfigMap, error) {
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).Patch(ctx, name, pt, data, opts, subresources...)
}

func (c *configMapClient) Apply(ctx context.Context, configMap *applycorev1.ConfigMapApplyConfiguration, opts metav1.ApplyOptions) (*v1.ConfigMap, error) {
	if err := c.init(); err != nil {
		return nil, err
	}
	return c.kc.CoreV1().ConfigMaps(c.namespace).Apply(ctx, configMap, opts)
}
