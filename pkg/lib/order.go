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

package lib

import (
	"encoding/json"
	"time"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/client/clientset/versioned"
	"kubepack.dev/lib-helm/action"
	"kubepack.dev/lib-helm/repo"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
	"gomodules.xyz/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func CreateOrder(reg *repo.Registry, bv v1alpha1.BundleView) (*v1alpha1.Order, error) {
	selection, err := toPackageSelection(reg, &bv.BundleOptionView, bv.LicenseKey)
	if err != nil {
		return nil, err
	}
	out := v1alpha1.Order{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.ResourceKindOrder,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              bv.Name,
			UID:               types.UID(uuid.New().String()),
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: v1alpha1.OrderSpec{
			Packages: selection,
		},
	}
	return &out, nil
}

// releaseNameMaxLen is the maximum length of a release name.
//
// As of Kubernetes 1.4, the max limit on a name is 63 chars. We reserve 10 for
// charts to add data. Effectively, that gives us 53 chars.
// See https://github.com/helm/helm/issues/1528
// xref: helm.sh/helm/v3/pkg/action/install.go
const releaseNameMaxLen = 53

func toPackageSelection(reg *repo.Registry, in *v1alpha1.BundleOptionView, licenseKey string) ([]v1alpha1.PackageSelection, error) {
	var out []v1alpha1.PackageSelection

	_, bundle, err := GetBundle(reg, &v1alpha1.BundleOption{
		BundleRef: v1alpha1.BundleRef{
			URL:  in.URL,
			Name: in.Name,
		},
		Version: in.Version,
	})
	if err != nil {
		return nil, err
	}

	for _, pkg := range in.Packages {
		if pkg.Chart != nil {
			if !pkg.Chart.Required {
				continue
			}

			for _, v := range pkg.Chart.Versions {
				if v.Selected {
					crds, waitFors, licenseKeyPath := FindChartData(bundle, pkg.Chart.ChartRef, v.Version)

					releaseName := pkg.Chart.Name
					if pkg.Chart.MultiSelect {
						releaseName += "-" + v.Version
					}
					if len(releaseName) > releaseNameMaxLen {
						return nil, errors.Errorf("release name %q exceeds max length of %d", releaseName, releaseNameMaxLen)
					}

					if len(licenseKeyPath) > 0 {
						var patch []jsonpatch.Operation
						if v.ValuesPatch != nil {
							err := json.Unmarshal(v.ValuesPatch.Raw, &patch)
							if err != nil {
								return nil, err
							}
						}
						injectLicenseKey := jsonpatch.NewOperation("replace", licenseKeyPath, licenseKey)
						patch = append(patch, injectLicenseKey)

						data, err := json.Marshal(patch)
						if err != nil {
							return nil, err
						}
						v.ValuesPatch = &runtime.RawExtension{Raw: data}
					}

					selection := v1alpha1.PackageSelection{
						Chart: &v1alpha1.ChartSelection{
							ChartRef:    pkg.Chart.ChartRef,
							Version:     v.Version,
							ReleaseName: releaseName,
							Namespace:   pkg.Chart.Namespace,
							ValuesFile:  v.ValuesFile,
							ValuesPatch: v.ValuesPatch,
							Resources:   crds,
							WaitFors:    waitFors,
							Bundle: &v1alpha1.ChartRepoRef{
								Name:    in.Name,
								URL:     in.URL,
								Version: in.Version,
							},
						},
					}
					out = append(out, selection)
				}
			}
		} else if pkg.Bundle != nil {
			selections, err := toPackageSelection(reg, pkg.Bundle, licenseKey)
			if err != nil {
				return nil, err
			}
			out = append(out, selections...)
		} else if pkg.OneOf != nil {
			return nil, errors.New("User must select one bundle")
		}
	}

	return out, nil
}

func FindChartData(bundle *v1alpha1.Bundle, chrtRef v1alpha1.ChartRef, chrtVersion string) (*v1alpha1.ResourceDefinitions, []v1alpha1.WaitFlags, string) {
	for _, pkg := range bundle.Spec.Packages {
		if pkg.Chart != nil &&
			pkg.Chart.URL == chrtRef.URL &&
			pkg.Chart.Name == chrtRef.Name {

			for _, v := range pkg.Chart.Versions {
				if v.Version == chrtVersion {
					return v.Resources, v.WaitFors, v.LicenseKeyPath
				}
			}
		}
	}
	return nil, nil, ""
}

func InstallOrder(getter genericclioptions.RESTClientGetter, reg *repo.Registry, order v1alpha1.Order) error {
	config, err := getter.ToRESTConfig()
	if err != nil {
		return err
	}

	kc, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	info, err := kc.ServerVersion()
	if err != nil {
		return err
	}
	kv, err := version.NewVersion(info.GitVersion)
	if err != nil {
		return err
	}
	kubeVersion := kv.ToMutator().ResetPrerelease().ResetMetadata().Done().String()

	namespaces := sets.NewString("default", "kube-system")

	f1 := &ApplicationCRDRegistrar{
		Config: config,
	}
	err = f1.Do()
	if err != nil {
		return err
	}

	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		if !namespaces.Has(pkg.Chart.Namespace) {
			f2 := &NamespaceCreator{
				Namespace: pkg.Chart.Namespace,
				Client:    kc,
			}
			err = f2.Do()
			if err != nil {
				return err
			}
			namespaces.Insert(pkg.Chart.Namespace)
		}

		f3, err := action.NewInstaller(getter, pkg.Chart.Namespace, "secret")
		if err != nil {
			return err
		}
		f3.
			WithRegistry(reg).
			WithOptions(action.InstallOptions{
				ChartURL:    pkg.Chart.ChartRef.URL,
				ChartName:   pkg.Chart.ChartRef.Name,
				Version:     pkg.Chart.Version,
				ValuesFile:  pkg.Chart.ValuesFile,
				ValuesPatch: pkg.Chart.ValuesPatch,
				Namespace:   pkg.Chart.Namespace,
				ReleaseName: pkg.Chart.ReleaseName,
			})
		err = f3.Do()
		if err != nil {
			return err
		}

		f4 := &WaitForChecker{
			Namespace:    pkg.Chart.Namespace,
			WaitFors:     pkg.Chart.WaitFors,
			ClientGetter: getter,
		}
		err = f4.Do()
		if err != nil {
			return err
		}

		if pkg.Chart.Resources != nil && len(pkg.Chart.Resources.Owned) > 0 {
			f5 := &CRDReadinessChecker{
				CRDs:   pkg.Chart.Resources.Owned,
				Client: kc.RESTClient(),
			}
			err = f5.Do()
			if err != nil {
				return err
			}
		}

		f6 := &ApplicationGenerator{
			Registry:    reg,
			Chart:       *pkg.Chart,
			KubeVersion: kubeVersion,
		}
		err = f6.Do()
		if err != nil {
			return err
		}

		f7 := &ApplicationCreator{
			App:    f6.Result(),
			Client: versioned.NewForConfigOrDie(config),
		}
		err = f7.Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func UninstallOrder(getter genericclioptions.RESTClientGetter, order v1alpha1.Order) error {
	for _, pkg := range order.Spec.Packages {
		if pkg.Chart == nil {
			continue
		}

		f1, err := action.NewUninstaller(getter, pkg.Chart.Namespace, "secret")
		if err != nil {
			return err
		}
		f1.WithReleaseName(pkg.Chart.ReleaseName)
		err = f1.Do()
		if err != nil {
			return err
		}
	}
	return nil
}
