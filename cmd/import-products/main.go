package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
			log.Fatal(err)
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
		nu.Spec.Phase = v1alpha1.ProductActive
		nu.Spec.Description = old.Description["markdown"]

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
		nu.Spec.SubProjects = map[string]v1alpha1.ProjectRef{}
		for k, v := range old.SubProjects {
			nu.Spec.SubProjects[k] = convertProjectRef(v)
		}

		data, err = json.MarshalIndent(&nu, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		f := filepath.Join("artifacts", "products", filename)
		err = os.MkdirAll(filepath.Dir(f), 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(f, data, 0644)
		if err != nil {
			log.Fatal(err)
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
		Version:  in.Version,
		HostDocs: in.HostDocs,
		Show:     in.Show,
		DocsDir:  in.DocsDir,
		Branch:   in.Branch,
		Info:     in.Info,
	}
}

func convertProjectRef(in api.ProjectRef) v1alpha1.ProjectRef {
	out := v1alpha1.ProjectRef{
		Dir: in.Dir,
	}
	for _, m := range in.Mappings {
		out.Mappings = append(out.Mappings, v1alpha1.Mapping{
			Versions:           m.Versions,
			SubProjectVersions: m.SubProjectVersions,
		})
	}
	return out
}
