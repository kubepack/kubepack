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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kubepack.dev/kubepack/apis"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/product"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func main() {
	stripe.Key = os.Getenv(apis.StripeAPIKey)

	dir := "artifacts/products"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		klog.Fatalln(err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filename := filepath.Join(dir, file.Name())
		fmt.Println(filename)

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			klog.Fatalln(err)
		}

		var existing v1alpha1.Product
		err = json.Unmarshal(data, &existing)
		if err != nil {
			klog.Fatalln(err)
		}
		if existing.Spec.UnitLabel == "" {
			existing.Spec.UnitLabel = "Cluster"
		}

		{
			exists := existing.Spec.StripeID != ""
			if exists {
				_, err := product.Get(existing.Spec.StripeID, nil)
				if err != nil {
					if se, ok := err.(*stripe.Error); ok && se.Code == stripe.ErrorCodeResourceMissing {
						exists = false
					} else {
						klog.Fatalln(err)
					}
				}
			}

			if !exists {
				params := &stripe.ProductParams{
					Name:        stripe.String(existing.Spec.Name),
					Type:        stripe.String(string(stripe.ProductTypeService)),
					Description: stripe.String(existing.Spec.Tagline),
					Active:      stripe.Bool(existing.Spec.Phase == v1alpha1.PhaseActive || existing.Spec.Phase == v1alpha1.PhaseArchived),
					Params: stripe.Params{
						Metadata: map[string]string{
							apis.StripeMetadataKeyUserID: strconv.FormatInt(existing.Spec.Owner, 10),
						},
					},
				}
				if existing.Spec.UnitLabel != "" {
					params.UnitLabel = stripe.String(existing.Spec.UnitLabel)
				}

				p, err := product.New(params)
				if err != nil {
					klog.Fatalln(err)
				}
				existing.Spec.StripeID = p.ID
			}
		}

		data, err = json.MarshalIndent(existing, "", "  ")
		if err != nil {
			klog.Fatalln(err)
		}
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			klog.Fatalln(err)
		}

		{
			planDir := filepath.Join(dir, existing.Name+"-plans")
			err = os.MkdirAll(planDir, 0755)
			if err != nil {
				klog.Fatalln(err)
			}
			plaanfiles, err := ioutil.ReadDir(planDir)
			if err != nil {
				klog.Fatalln(err)
			}
			if len(plaanfiles) == 0 {
				params := &stripe.PlanParams{
					Nickname:  stripe.String(existing.Spec.ShortName + " Community"),
					ProductID: stripe.String(existing.Spec.StripeID),
					Amount:    stripe.Int64(0),
					Interval:  stripe.String(string(stripe.PlanIntervalMonth)),
					Currency:  stripe.String(string(stripe.CurrencyUSD)),
				}
				p, err := plan.New(params)
				if err != nil {
					klog.Fatalln(err)
				}
				plaan := v1alpha1.Plan{
					ObjectMeta: metav1.ObjectMeta{
						Name: strings.ReplaceAll(strings.ToLower(stripe.StringValue(params.Nickname)), " ", "-"),
					},
					Spec: v1alpha1.PlanSpec{
						StripeID:  p.ID,
						NickName:  stripe.StringValue(params.Nickname),
						ProductID: existing.Spec.StripeID,
						Bundle: &v1alpha1.ChartRef{
							URL:  apis.BundleRepoURL,
							Name: existing.Spec.Key + "-community",
						},
						Phase:         v1alpha1.PhaseActive,
						IncludedPlans: nil,
						//Amount:        stripe.Int64(p.Amount),
						Interval: stripe.String(string(p.Interval)),
						Currency: stripe.String(string(p.Currency)),
					},
				}

				data, err = json.MarshalIndent(plaan, "", "  ")
				if err != nil {
					klog.Fatalln(err)
				}
				err = ioutil.WriteFile(filepath.Join(planDir, plaan.Name+".json"), data, 0644)
				if err != nil {
					klog.Fatalln(err)
				}
			} else {
				for _, plaanfile := range plaanfiles {
					if plaanfile.IsDir() {
						continue
					}
					data, err := ioutil.ReadFile(filepath.Join(planDir, plaanfile.Name()))
					if err != nil {
						klog.Fatalln(err)
					}
					var plaan v1alpha1.Plan
					err = json.Unmarshal(data, &plaan)
					if err != nil {
						klog.Fatalln(err)
					}

					plaan.Spec.ProductID = existing.Spec.StripeID

					exists := plaan.Spec.StripeID != ""
					if exists {
						_, err := plan.Get(plaan.Spec.StripeID, nil)
						if err != nil {
							if se, ok := err.(*stripe.Error); ok && se.Code == stripe.ErrorCodeResourceMissing {
								exists = false
							} else {
								klog.Fatalln(err)
							}
						}
					}
					if !exists {
						params := &stripe.PlanParams{
							Nickname:  stripe.String(plaan.Spec.NickName),
							ProductID: stripe.String(existing.Spec.StripeID),

							AggregateUsage:  plaan.Spec.AggregateUsage,
							Amount:          plaan.Spec.Amount,
							AmountDecimal:   plaan.Spec.AmountDecimal,
							BillingScheme:   plaan.Spec.BillingScheme,
							Currency:        plaan.Spec.Currency,
							Interval:        plaan.Spec.Interval,
							IntervalCount:   plaan.Spec.IntervalCount,
							Tiers:           convertPlanTier(plaan.Spec.Tiers),
							TiersMode:       plaan.Spec.TiersMode,
							TrialPeriodDays: plaan.Spec.TrialPeriodDays,
							UsageType:       plaan.Spec.UsageType,
						}
						if plaan.Spec.TransformUsage != nil {
							params.TransformUsage = &stripe.PlanTransformUsageParams{
								DivideBy: plaan.Spec.TransformUsage.DivideBy,
								Round:    plaan.Spec.TransformUsage.Round,
							}
						}
						p, err := plan.New(params)
						if err != nil {
							klog.Fatalln(err)
						}

						plaan.Spec.StripeID = p.ID
						plaan.Name = strings.ReplaceAll(strings.ToLower(stripe.StringValue(params.Nickname)), " ", "-")

						data, err = json.MarshalIndent(plaan, "", "  ")
						if err != nil {
							klog.Fatalln(err)
						}
						err = ioutil.WriteFile(filepath.Join(planDir, plaan.Name+".json"), data, 0644)
						if err != nil {
							klog.Fatalln(err)
						}
					}
				}
			}
		}
	}
}

func convertPlanTier(in []*v1alpha1.PlanTier) []*stripe.PlanTierParams {
	var out []*stripe.PlanTierParams
	for _, tier := range in {
		out = append(out, &stripe.PlanTierParams{
			FlatAmount:        tier.FlatAmount,
			FlatAmountDecimal: tier.FlatAmountDecimal,
			UnitAmount:        tier.UnitAmount,
			UnitAmountDecimal: tier.UnitAmountDecimal,
			UpTo:              tier.UpTo,
		})
	}
	return out
}
