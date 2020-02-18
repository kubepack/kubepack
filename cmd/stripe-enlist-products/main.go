package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"kubepack.dev/kubepack/apis"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/product"
)

func main() {
	stripe.Key = os.Getenv(apis.StripeAPIKey)

	dir := "artifacts/products"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := filepath.Join(dir, file.Name())
		fmt.Println(filename)

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalln(err)
		}

		var existing v1alpha1.Product
		err = json.Unmarshal(data, &existing)
		if err != nil {
			log.Fatalln(err)
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
						log.Fatalln(err)
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
					log.Fatalln(err)
				}
				existing.Spec.StripeID = p.ID
			}
		}

		if len(existing.Spec.Plans) == 0 {
			params := &stripe.PlanParams{
				Nickname:  stripe.String(existing.Spec.ShortName + " Community"),
				ProductID: stripe.String(existing.Spec.StripeID),
				Amount:    stripe.Int64(0),
				Interval:  stripe.String(string(stripe.PlanIntervalMonth)),
				Currency:  stripe.String(string(stripe.CurrencyUSD)),
			}
			p, err := plan.New(params)
			if err != nil {
				log.Fatalln(err)
			}
			existing.Spec.Plans = []v1alpha1.Plan{
				{
					StripeID: p.ID,
					NickName: stripe.StringValue(params.Nickname),
					Chart: v1alpha1.ChartRef{
						URL:  apis.BundleRepoURL,
						Name: existing.Spec.Key + "-community",
					},
					Phase:         v1alpha1.PhaseActive,
					IncludedPlans: nil,
					Amount:        p.Amount,
					Interval:      p.Interval,
					Currency:      p.Currency,
				},
			}
		} else {
			for idx, plaan := range existing.Spec.Plans {
				exists := plaan.StripeID != ""
				if exists {
					_, err := plan.Get(plaan.StripeID, nil)
					if err != nil {
						if se, ok := err.(*stripe.Error); ok && se.Code == stripe.ErrorCodeResourceMissing {
							exists = false
						} else {
							log.Fatalln(err)
						}
					}
				}
				if !exists {
					params := &stripe.PlanParams{
						Nickname:  stripe.String(plaan.NickName),
						ProductID: stripe.String(existing.Spec.StripeID),

						AggregateUsage: StringP(plaan.AggregateUsage),
						Amount:         Int64P(plaan.Amount),
						AmountDecimal:  Float64P(plaan.AmountDecimal),
						BillingScheme:  StringP(string(plaan.BillingScheme)),
						Currency:       StringP(string(plaan.Currency)),
						Interval:       StringP(string(plaan.Interval)),
						IntervalCount:  Int64P(plaan.IntervalCount),
						Tiers:          convertPlanTier(plaan.Tiers),
						TiersMode:      StringP(plaan.TiersMode),
						TransformUsage: &stripe.PlanTransformUsageParams{
							DivideBy: Int64P(plaan.TransformUsage.DivideBy),
							Round:    StringP(string(plaan.TransformUsage.Round)),
						},
						TrialPeriodDays: Int64P(plaan.TrialPeriodDays),
						UsageType:       StringP(string(plaan.UsageType)),
					}
					p, err := plan.New(params)
					if err != nil {
						log.Fatalln(err)
					}
					existing.Spec.Plans[idx].StripeID = p.ID
				}
			}
		}

		data, err = json.MarshalIndent(existing, "", "  ")
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func convertPlanTier(in []*v1alpha1.PlanTier) []*stripe.PlanTierParams {
	var out []*stripe.PlanTierParams
	for _, tier := range in {
		out = append(out, &stripe.PlanTierParams{
			Params:            stripe.Params{},
			FlatAmount:        Int64P(tier.FlatAmount),
			FlatAmountDecimal: Float64P(tier.FlatAmountDecimal),
			UnitAmount:        Int64P(tier.UnitAmount),
			UnitAmountDecimal: Float64P(tier.UnitAmountDecimal),
			UpTo:              Int64P(tier.UpTo),
		})
	}
	return out
}

func StringP(v string) *string {
	if v != "" {
		return &v
	}
	return nil
}

func Int64P(v int64) *int64 {
	if v != 0 {
		return &v
	}
	return nil
}
func Float64P(v float64) *float64 {
	if v != 0 {
		return &v
	}
	return nil
}
