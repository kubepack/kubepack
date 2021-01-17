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
	"fmt"
	"os"

	"kubepack.dev/kubepack/apis"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/product"
)

func main() {
	stripe.Key = os.Getenv(apis.StripeAPIKey)

	{
		var entries []string

		params := &stripe.PlanListParams{}
		params.Filters.AddFilter("limit", "", "3")
		i := plan.List(params)
		for i.Next() {
			p := i.Plan()
			fmt.Println(p.Nickname)
			entries = append(entries, p.ID)
		}

		for _, id := range entries {
			_, _ = plan.Del(id, nil)
		}
	}

	{
		var entries []string

		params := &stripe.ProductListParams{}
		params.Filters.AddFilter("limit", "", "3")
		i := product.List(params)
		for i.Next() {
			p := i.Product()
			fmt.Println(p.Name)
			entries = append(entries, p.ID)
		}

		for _, id := range entries {
			_, _ = product.Del(id, nil)
		}
	}
}
