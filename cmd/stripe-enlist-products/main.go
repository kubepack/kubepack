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

			data, err := json.MarshalIndent(existing, "", "  ")
			if err != nil {
				log.Fatalln(err)
			}
			err = ioutil.WriteFile(filename, data, 0644)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}
