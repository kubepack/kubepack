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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/klog/v2"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/appscode/static-assets/api"
	"github.com/appscode/static-assets/data/products"
)

func main() {
	for _, filename := range products.AssetNames() {
		data := products.MustAsset(filename)
		var old api.Product
		err := json.Unmarshal(data, &old)
		if err != nil {
			klog.Fatal(err)
		}

		if !old.Published {
			continue
		}

		var nu v1alpha1.Product

		nu.Name = old.Key

		nu.Spec.StripeID = old.StripeProductID
		nu.Spec.Key = old.Key
		nu.Spec.Name = old.Fullname
		nu.Spec.ShortName = old.Name
		nu.Spec.Tagline = old.Tagline
		nu.Spec.Summary = old.Summary
		nu.Spec.Owner = 1
		nu.Spec.Phase = v1alpha1.PhaseActive
		nu.Spec.Description = old.Description["html"]

		media := map[v1alpha1.MediaType]v1alpha1.MediaSpec{}
		if old.HeroImage.Src != "" {
			media[v1alpha1.MediaHeroImage] = v1alpha1.MediaSpec{
				Description: v1alpha1.MediaHeroImage,
				ImageSpec: v1alpha1.ImageSpec{
					Source: old.HeroImage.Src,
				},
			}
		}
		if old.Logo.Src != "" {
			media[v1alpha1.MediaLogo] = v1alpha1.MediaSpec{
				Description: v1alpha1.MediaLogo,
				ImageSpec: v1alpha1.ImageSpec{
					Source: old.Logo.Src,
				},
			}
		}
		if old.LogoWhite.Src != "" {
			media[v1alpha1.MediaLogoWhite] = v1alpha1.MediaSpec{
				Description: v1alpha1.MediaLogoWhite,
				ImageSpec: v1alpha1.ImageSpec{
					Source: old.LogoWhite.Src,
				},
			}
		}
		if old.Icon.Src != "" {
			media[v1alpha1.MediaIcon] = v1alpha1.MediaSpec{
				Description: v1alpha1.MediaIcon,
				ImageSpec: v1alpha1.ImageSpec{
					Source: old.Icon.Src,
				},
			}
			for k, v := range old.Icon.Sizes {
				t := v1alpha1.MediaType("icon_" + k)
				media[t] = v1alpha1.MediaSpec{
					Description: t,
					ImageSpec: v1alpha1.ImageSpec{
						Source: v,
					},
				}
			}
		}

		for k := range media {
			nu.Spec.Media = append(nu.Spec.Media, media[k])
		}

		nu.Spec.Maintainers = []v1alpha1.ContactData{
			{
				Name:  old.Author,
				Email: "support@appscode.com",
			},
		}

		if old.Keywords != "" {
			keywords := strings.Split(old.Keywords, " ")
			if len(keywords) > 0 {
				nu.Spec.Keywords = keywords
			}
		}

		links := map[v1alpha1.LinkType]v1alpha1.Link{}
		if old.RepoURL != "" {
			links[v1alpha1.LinkSourceRepo] = v1alpha1.Link{
				Description: v1alpha1.LinkSourceRepo,
				URL:         old.RepoURL,
			}
		}
		if old.StarRepo != "" {
			links[v1alpha1.LinkStarRepo] = v1alpha1.Link{
				Description: v1alpha1.LinkStarRepo,
				URL:         old.StarRepo,
			}
		}
		if old.DocRepo != "" {
			links[v1alpha1.LinkDocsRepo] = v1alpha1.Link{
				Description: v1alpha1.LinkDocsRepo,
				URL:         old.DocRepo,
			}
		}
		if old.DatasheetFormID != "" {
			links[v1alpha1.LinkDatasheetFormID] = v1alpha1.Link{
				Description: v1alpha1.LinkDatasheetFormID,
				URL:         old.DatasheetFormID,
			}
		}
		for k, v := range old.SocialLinks {
			t := v1alpha1.LinkType(strings.ToLower(k))
			links[t] = v1alpha1.Link{
				Description: t,
				URL:         v,
			}
		}
		for k, v := range old.SupportLinks {
			var t v1alpha1.LinkType
			switch k {
			case "Support URL":
				t = v1alpha1.LinkSupportDesk
			case "Website URL":
				t = v1alpha1.LinkWebsite
			default:
				t = v1alpha1.LinkType(strings.ToLower(k))
			}
			links[t] = v1alpha1.Link{
				Description: t,
				URL:         v,
			}
		}
		// add links in sorted order of link type
		lt := make([]v1alpha1.LinkType, 0, len(links))
		for k := range links {
			lt = append(lt, k)
		}
		sort.Slice(lt, func(i, j int) bool { return lt[i] < lt[j] })
		nu.Spec.Links = make([]v1alpha1.Link, 0, len(links))
		for _, t := range lt {
			nu.Spec.Links = append(nu.Spec.Links, links[t])
		}

		for _, in := range old.Badges {
			nu.Spec.Badges = append(nu.Spec.Badges, convertBadge(in))
		}
		// Plans
		for _, in := range old.Versions {
			nu.Spec.Versions = append(nu.Spec.Versions, convertVersion(in))
		}
		nu.Spec.LatestVersion = old.LatestVersion

		f := filepath.Join("artifacts", "products", filename)
		err = os.MkdirAll(filepath.Dir(f), 0755)
		if err != nil {
			klog.Fatal(err)
		}

		if _, err := os.Stat(f); err == nil {
			data, err := ioutil.ReadFile(f)
			if err != nil {
				klog.Fatal(err)
			}

			var existing v1alpha1.Product
			err = json.Unmarshal(data, &existing)
			if err != nil {
				klog.Fatal(err)
			}
			// preserve product id and plans
			nu.Spec.StripeID = existing.Spec.StripeID
		}

		data, err = json.MarshalIndent(&nu, "", "  ")
		if err != nil {
			klog.Fatal(err)
		}
		err = ioutil.WriteFile(f, data, 0644)
		if err != nil {
			klog.Fatal(err)
		}
	}
}

func convertBadge(in api.Badge) v1alpha1.Badge {
	return v1alpha1.Badge{
		URL:  in.URL,
		Alt:  in.Alt,
		Logo: in.Logo,
	}
}

func convertVersion(in api.ProductVersion) v1alpha1.ProductVersion {
	return v1alpha1.ProductVersion{
		Version: in.Version,
	}
}
