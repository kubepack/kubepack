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

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"
	"kubepack.dev/lib-helm/pkg/repo"

	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var (
	// url     = "https://charts.appscode.com/stable/"
	// name    = "kubedb"
	// version = "v0.13.0-rc.0"

	url     = "https://kubernetes-charts.storage.googleapis.com"
	name    = "wordpress"
	version = ""
)

func main() {
	flag.StringVar(&url, "url", url, "Chart repo url")
	flag.StringVar(&name, "name", name, "Name of bundle")
	flag.StringVar(&version, "version", version, "Version of bundle")
	flag.Parse()

	chrt, err := lib.DefaultRegistry.GetChart(url, name, version)
	if err != nil {
		klog.Fatalln(err)
	}

	repoName, err := repo.DefaultNamer.Name(url)
	if err != nil {
		klog.Fatalln(err)
	}

	var nu v1alpha1.Product
	nu.Name = repoName + "-" + chrt.Name()

	nu.Spec.StripeID = ""
	nu.Spec.Key = nu.Name
	nu.Spec.Name = chrt.Name()
	nu.Spec.ShortName = chrt.Name()
	nu.Spec.Tagline = chrt.Metadata.Description
	nu.Spec.Summary = ""
	nu.Spec.Owner = -1
	nu.Spec.Description = chrt.Metadata.Description
	nu.Spec.UnitLabel = ""
	if chrt.Removed {
		nu.Spec.Phase = v1alpha1.PhaseArchived
	} else {
		nu.Spec.Phase = v1alpha1.PhaseActive
	}
	nu.Spec.Media = []v1alpha1.MediaSpec{
		{
			Description: v1alpha1.MediaIcon,
			ImageSpec: v1alpha1.ImageSpec{
				Source:    chrt.Metadata.Icon,
				TotalSize: "",
				Type:      "",
			},
		},
	}
	for _, maintainer := range chrt.Metadata.Maintainers {
		nu.Spec.Maintainers = append(nu.Spec.Maintainers, v1alpha1.ContactData{
			Name:  maintainer.Name,
			URL:   maintainer.URL,
			Email: maintainer.Email,
		})
	}
	nu.Spec.Keywords = chrt.Metadata.Keywords
	if chrt.Metadata.Home != "" {
		nu.Spec.Links = []v1alpha1.Link{
			{
				Description: v1alpha1.LinkWebsite,
				URL:         chrt.Metadata.Home,
			},
		}
	}
	rlDate := metav1.NewTime(chrt.Created)
	nu.Spec.Versions = []v1alpha1.ProductVersion{
		{
			Version:     chrt.Metadata.Version,
			ReleaseDate: &rlDate,
		},
	}
	nu.Spec.LatestVersion = chrt.Metadata.Version // Not AppVersion

	/*
		nu.Spec.Plans = []v1alpha1.Plan{
			{
				StripeID:        "",
				NickName:        "",
				Chart:           v1alpha1.ChartRef{},
				Phase:           "",
				IncludedPlans:   nil,
				AggregateUsage:  "",
				Amount:          0,
				AmountDecimal:   0,
				BillingScheme:   "",
				Currency:        "",
				Interval:        "",
				IntervalCount:   0,
				Tiers:           nil,
				TiersMode:       "",
				TransformUsage:  nil,
				TrialPeriodDays: 0,
				UsageType:       "",
			},
		}
	*/

	data, err := json.MarshalIndent(nu, "", "  ")
	if err != nil {
		klog.Fatalln(err)
	}
	fmt.Println(string(data))
}
