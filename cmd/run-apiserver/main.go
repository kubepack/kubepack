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

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"time"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/artifacts/products"
	"kubepack.dev/kubepack/pkg/handler"
	"kubepack.dev/kubepack/pkg/lib"
	"kubepack.dev/lib-helm/getter"
	"kubepack.dev/lib-helm/repo"

	gomime "github.com/cubewise-code/go-mime"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-macaron/binding"
	"github.com/google/uuid"
	"github.com/spf13/pflag"
	"gopkg.in/macaron.v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/util/homedir"
	cliflag "k8s.io/component-base/cli/flag"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"kmodules.xyz/client-go/logs"
	"kmodules.xyz/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

func main() {
	flags := pflag.CommandLine
	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(flags)

	flags.AddGoFlagSet(flag.CommandLine)
	logs.ParseFlags()

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer())

	// PUBLIC
	m.Group("/bundleview", func() {
		m.Get("", binding.Json(v1alpha1.ChartRepoRef{}), func(ctx *macaron.Context, params v1alpha1.ChartRepoRef) {
			// TODO: verify params

			bv, err := lib.CreateBundleViewForChart(lib.DefaultRegistry, &params)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			ctx.JSON(http.StatusOK, bv)
		})

		// Generate Order for a BundleView
		m.Post("/orders", binding.Json(v1alpha1.BundleView{}), func(ctx *macaron.Context, params v1alpha1.BundleView) {
			order, err := lib.CreateOrder(lib.DefaultRegistry, params)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := yaml.Marshal(order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			err = bs.WriteFile(ctx.Req.Context(), path.Join(string(order.UID), "order.yaml"), data)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.JSON(http.StatusOK, order)
		})
	})

	// PUBLIC
	m.Group("/packageview", func() {
		m.Get("", binding.Json(v1alpha1.ChartRepoRef{}), func(ctx *macaron.Context, params v1alpha1.ChartRepoRef) {
			// TODO: verify params

			chrt, err := lib.DefaultRegistry.GetChart(params.URL, params.Name, params.Version)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			pv, err := lib.CreatePackageView(params.URL, chrt.Chart)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			ctx.JSON(http.StatusOK, pv)
		})

		// Generate Order for a Editor PackageView / Chart
		m.Post("/orders", binding.Json(lib.ChartOrder{}), func(ctx *macaron.Context, params lib.ChartOrder) {
			order, err := lib.CreateChartOrder(lib.DefaultRegistry, params)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := yaml.Marshal(order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			err = bs.WriteFile(ctx.Req.Context(), path.Join(string(order.UID), "order.yaml"), data)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.JSON(http.StatusOK, order)
		})
	})

	// PUBLIC
	m.Get("/packageview/files", binding.Json(v1alpha1.ChartRepoRef{}), func(ctx *macaron.Context, params v1alpha1.ChartRepoRef) {
		// TODO: verify params

		chrt, err := lib.DefaultRegistry.GetChart(params.URL, params.Name, params.Version)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		files := make([]string, 0, len(chrt.Raw))
		for _, f := range chrt.Raw {
			files = append(files, f.Name)
		}
		sort.Strings(files)

		ctx.JSON(http.StatusOK, files)
	})

	// PUBLIC
	m.Get("/packageview/files/*", binding.Json(v1alpha1.ChartRepoRef{}), func(ctx *macaron.Context, params v1alpha1.ChartRepoRef) {
		// TODO: verify params

		chrt, err := lib.DefaultRegistry.GetChart(params.URL, params.Name, params.Version)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		filename := ctx.Params("*")
		for _, f := range chrt.Raw {
			if f.Name == filename {
				ext := filepath.Ext(f.Name)
				if ct := gomime.TypeByExtension(ext); ct != "" {
					ctx.Header().Set("Content-Type", ct)
				} else {
					ct := mimetype.Detect(f.Data)
					ctx.Header().Set("Content-Type", ct.String())
				}
				_, _ = ctx.Write(f.Data)
				return
			}
		}

		ctx.WriteHeader(http.StatusNotFound)
	})

	m.Group("/editor/:group/:version/:resource", func() {
		// PUBLIC
		// INITIAL Model (Values)
		// GET vs POST (Get makes more sense, but do we send so much data via query string?)
		// With POST, we can send large payloads without any non-standard limits
		// https://stackoverflow.com/a/812962
		m.Post("/model", binding.Json(unstructured.Unstructured{}), func(ctx *macaron.Context, opts unstructured.Unstructured) {
			gvr := schema.GroupVersionResource{
				Group:    ctx.Params(":group"),
				Version:  ctx.Params(":version"),
				Resource: ctx.Params(":resource"),
			}

			model, err := lib.GenerateEditorModel(lib.DefaultRegistry, gvr, opts)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.JSON(http.StatusOK, model)
		})
		m.Post("/:releaseName/namespaces/:namespace/manifest", binding.Json(unstructured.Unstructured{}), func(ctx *macaron.Context, opts unstructured.Unstructured) {
			gvr := schema.GroupVersionResource{
				Group:    ctx.Params(":group"),
				Version:  ctx.Params(":version"),
				Resource: ctx.Params(":resource"),
			}
			rlm := lib.ReleaseMetadata{
				Name:      ctx.Params(":releaseName"),
				Namespace: ctx.Params(":namespace"),
			}
			manifest, _, err := lib.RenderChartTemplate(lib.DefaultRegistry, gvr, rlm, opts)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			_, _ = ctx.Write([]byte(manifest))
		})
		m.Post("/:releaseName/namespaces/:namespace/resources", binding.Json(unstructured.Unstructured{}), func(ctx *macaron.Context, opts unstructured.Unstructured) {
			gvr := schema.GroupVersionResource{
				Group:    ctx.Params(":group"),
				Version:  ctx.Params(":version"),
				Resource: ctx.Params(":resource"),
			}
			rlm := lib.ReleaseMetadata{
				Name:      ctx.Params(":releaseName"),
				Namespace: ctx.Params(":namespace"),
			}
			_, tpls, err := lib.RenderChartTemplate(lib.DefaultRegistry, gvr, rlm, opts)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			if ctx.QueryBool("skipCRDs") {
				tpls.CRDs = nil
			}
			ctx.JSON(http.StatusOK, tpls)
		})
	})

	// PUBLIC
	m.Get("/products", func(ctx *macaron.Context) {
		// /products

		phase := ctx.Params(":phase")
		var out v1alpha1.ProductList
		for _, filename := range products.AssetNames() {
			data, err := products.Asset(filename)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			var p v1alpha1.Product
			err = json.Unmarshal(data, &p)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			if phase == "" || p.Spec.Phase == v1alpha1.Phase(phase) {
				out.Items = append(out.Items, p)
			}
		}
		ctx.JSON(http.StatusOK, out)
	})

	// PUBLIC
	m.Get("/products/:owner/:key", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		data, err := products.Asset(ctx.Params(":key") + ".json")
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		_, _ = ctx.Write(data)
	})

	// PUBLIC
	m.Get("/products/:owner/:key/plans", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		dir := "artifacts/products/kubedb-plans"
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		phase := ctx.Params(":phase")
		var out v1alpha1.PlanList
		for _, file := range files {
			data, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			var plaan v1alpha1.Plan
			err = json.Unmarshal(data, &plaan)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			if phase == "" || plaan.Spec.Phase == v1alpha1.Phase(phase) {
				out.Items = append(out.Items, plaan)
			}
		}
		ctx.JSON(http.StatusOK, out)
	})

	// PUBLIC
	m.Get("/products/:owner/:key/plans/:plan", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		dir := "artifacts/products/" + ctx.Params(":key") + "-plans"

		data, err := ioutil.ReadFile(filepath.Join(dir, ctx.Params(":plan")+".json"))
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		var plaan v1alpha1.Plan
		err = json.Unmarshal(data, &plaan)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, plaan)
	})

	// PUBLIC
	m.Get("/products/:owner/:key/compare", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		phase := ctx.Params(":phase")

		// product
		data, err := products.Asset(ctx.Params(":key") + ".json")
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		var p v1alpha1.Product
		err = json.Unmarshal(data, &p)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		var url string
		version := p.Spec.LatestVersion
		var plaans []v1alpha1.Plan

		dir := "artifacts/products/" + ctx.Params(":key") + "-plans"
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		for idx, file := range files {
			data, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			var plaan v1alpha1.Plan
			err = json.Unmarshal(data, &plaan)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			if idx == 0 {
				url = plaan.Spec.Bundle.URL
			}

			if phase == "" || plaan.Spec.Phase == v1alpha1.Phase(phase) {
				plaans = append(plaans, plaan)
			}
		}

		sort.Slice(plaans, func(i, j int) bool {
			if plaans[i].Spec.Weight == plaans[j].Spec.Weight {
				return plaans[i].Spec.NickName < plaans[j].Spec.NickName
			}
			return plaans[i].Spec.Weight < plaans[j].Spec.Weight
		})

		var names []string
		for _, plaan := range plaans {
			names = append(names, plaan.Spec.Bundle.Name)
		}

		table, err := lib.ComparePlans(lib.DefaultRegistry, url, names, version)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		table.Plans = plaans
		ctx.JSON(http.StatusOK, table)
	})

	// PUBLIC
	m.Get("/products/:owner/:key/plans/:plan/bundleview", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		// product
		data, err := products.Asset(ctx.Params(":key") + ".json")
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		var p v1alpha1.Product
		err = json.Unmarshal(data, &p)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		// plan
		dir := "artifacts/products/" + ctx.Params(":key") + "-plans"
		data, err = ioutil.ReadFile(filepath.Join(dir, ctx.Params(":plan")+".json"))
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		var plaan v1alpha1.Plan
		err = json.Unmarshal(data, &plaan)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		bv, err := lib.CreateBundleViewForBundle(lib.DefaultRegistry, &v1alpha1.ChartRepoRef{
			URL:     plaan.Spec.Bundle.URL,
			Name:    plaan.Spec.Bundle.Name,
			Version: p.Spec.LatestVersion,
		})
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, bv)
	})

	// PUBLIC
	m.Get("/product_id/:id", func(ctx *macaron.Context) {
		// TODO: get product by id

		data, err := products.Asset("kubedb.json")
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		_, _ = ctx.Write(data)
	})

	// PRIVATE
	// Should we store Order UID in a table per User?
	m.Group("/deploy/orders", func() {
		// Generate Order for a single chart and optional values patch
		m.Post("", binding.Json(v1alpha1.Order{}), func(ctx *macaron.Context, order v1alpha1.Order) {
			if len(order.Spec.Packages) == 0 {
				ctx.Error(http.StatusBadRequest, "missing package selection for order")
				return
			}

			order.TypeMeta = metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
				Kind:       v1alpha1.ResourceKindOrder,
			}
			order.ObjectMeta = metav1.ObjectMeta{
				UID:               types.UID(uuid.New().String()),
				CreationTimestamp: metav1.NewTime(time.Now()),
			}
			if order.Name == "" {
				order.Name = order.Spec.Packages[0].Chart.ReleaseName
			}

			data, err := yaml.Marshal(order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			err = bs.WriteFile(ctx.Req.Context(), path.Join(string(order.UID), "order.yaml"), data)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.JSON(http.StatusOK, order)
		})
		m.Get("/:id/render/manifest", func(ctx *macaron.Context) {
			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := bs.ReadFile(ctx.Req.Context(), path.Join(ctx.Params(":id"), "order.yaml"))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			manifest, _, err := lib.RenderOrderTemplate(bs, lib.DefaultRegistry, order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			_, _ = ctx.Write([]byte(manifest))
		})
		m.Get("/:id/render/resources", func(ctx *macaron.Context) {
			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := bs.ReadFile(ctx.Req.Context(), path.Join(ctx.Params(":id"), "order.yaml"))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			_, tpls, err := lib.RenderOrderTemplate(bs, lib.DefaultRegistry, order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.JSON(http.StatusOK, tpls)
		})
		m.Get("/:id/helm2", func(ctx *macaron.Context) {
			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := bs.ReadFile(ctx.Req.Context(), path.Join(ctx.Params(":id"), "order.yaml"))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			scripts, err := lib.GenerateHelm2Script(bs, lib.DefaultRegistry, order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.JSON(http.StatusOK, scripts)
		})
		m.Get("/:id/helm3", func(ctx *macaron.Context) {
			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := bs.ReadFile(ctx.Req.Context(), path.Join(ctx.Params(":id"), "order.yaml"))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			scripts, err := lib.GenerateHelm3Script(bs, lib.DefaultRegistry, order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			ctx.JSON(http.StatusOK, scripts)
		})
		m.Get("/:id/yaml", func(ctx *macaron.Context) {
			bs, err := lib.NewTestBlobStore()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := bs.ReadFile(ctx.Req.Context(), path.Join(ctx.Params(":id"), "order.yaml"))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			scripts, err := lib.GenerateYAMLScript(bs, lib.DefaultRegistry, order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			ctx.JSON(http.StatusOK, scripts)
		})
	})

	// PRIVATE
	m.Group("/clusters/:cluster", func() {
		m.Group("/editor/:group/:version/namespaces/:namespace/:resource/:releaseName", func() {
			// GET Model from Existing Installations
			m.Get("/model", func(ctx *macaron.Context) {
				cfg, err := clientcmd.BuildConfigFromContext("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}

				gvr := schema.GroupVersionResource{
					Group:    ctx.Params(":group"),
					Version:  ctx.Params(":version"),
					Resource: ctx.Params(":resource"),
				}
				opts := lib.ReleaseMetadata{
					Name:      ctx.Params(":releaseName"),
					Namespace: ctx.Params(":namespace"),
				}
				tpl, err := lib.LoadEditorModel(cfg, lib.DefaultRegistry, gvr, opts)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}
				data, err := yaml.Marshal(tpl.Values)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}
				_, _ = ctx.Write(data)
			})

			// redundant apis
			// can be replaced by getting the model, then using the /editor apis
			m.Get("/manifest", func(ctx *macaron.Context) {
				cfg, err := clientcmd.BuildConfigFromContext("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}

				gvr := schema.GroupVersionResource{
					Group:    ctx.Params(":group"),
					Version:  ctx.Params(":version"),
					Resource: ctx.Params(":resource"),
				}
				opts := lib.ReleaseMetadata{
					Name:      ctx.Params(":releaseName"),
					Namespace: ctx.Params(":namespace"),
				}
				tpl, err := lib.LoadEditorModel(cfg, lib.DefaultRegistry, gvr, opts)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}
				_, _ = ctx.Write(tpl.Manifest)
			})

			// redundant apis
			// can be replaced by getting the model, then using the /editor apis
			m.Get("/resources", func(ctx *macaron.Context) {
				cfg, err := clientcmd.BuildConfigFromContext("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}

				gvr := schema.GroupVersionResource{
					Group:    ctx.Params(":group"),
					Version:  ctx.Params(":version"),
					Resource: ctx.Params(":resource"),
				}
				opts := lib.ReleaseMetadata{
					Name:      ctx.Params(":releaseName"),
					Namespace: ctx.Params(":namespace"),
				}
				tpl, err := lib.LoadEditorModel(cfg, lib.DefaultRegistry, gvr, opts)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}

				out := struct {
					Resources []*unstructured.Unstructured `json:"resources,omitempty"`
				}{
					tpl.Resources,
				}
				data, err := yaml.Marshal(out)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}
				_, _ = ctx.Write(data)
			})

			// create / update / apply / install
			m.Put("", binding.Json(unstructured.Unstructured{}), func(ctx *macaron.Context, model unstructured.Unstructured) {
				rls, err := handler.ApplyResource(ctx, model, f)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}
				ctx.JSON(http.StatusOK, rls)
			})

			m.Delete("", func(ctx *macaron.Context) {
				rls, err := handler.DeleteResource(ctx, f)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err.Error())
					return
				}
				ctx.JSON(http.StatusOK, rls)
			})
		})

		m.Post("/deploy/:id", func(ctx *macaron.Context) {
		})
		m.Delete("/deploy/:id", func(ctx *macaron.Context) {
		})
	})

	m.Get("/chartrepositories", func(ctx *macaron.Context) {
		repos, err := repo.DefaultNamer.ListHelmHubRepositories()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, repos)
	})
	m.Get("/chartrepositories/charts", func(ctx *macaron.Context) {
		url := ctx.Query("url")
		if url == "" {
			ctx.Error(http.StatusBadRequest, "missing url")
			return
		}

		cfg, _, err := lib.DefaultRegistry.Get(url)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		cr, err := repo.NewChartRepository(cfg, getter.All())
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		err = cr.Load()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, cr.ListCharts())
	})
	m.Get("/chartrepositories/charts/:name/versions", func(ctx *macaron.Context) {
		url := ctx.Query("url")
		name := ctx.Params("name")

		if url == "" {
			ctx.Error(http.StatusBadRequest, "missing url")
			return
		}
		if name == "" {
			ctx.Error(http.StatusBadRequest, "missing chart name")
			return
		}

		cfg, _, err := lib.DefaultRegistry.Get(url)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		cr, err := repo.NewChartRepository(cfg, getter.All())
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		err = cr.Load()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, cr.ListVersions(name))
	})

	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Run()
}
