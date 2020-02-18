package main

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/product"
)

func main() {
	stripe.Key = os.Getenv("STRIPE_API_KEY")

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
