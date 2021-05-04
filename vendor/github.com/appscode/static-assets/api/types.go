package api

import (
	"encoding/json"
)

type Feature struct {
	Title       string `json:"title"`
	Image       Image  `json:"image"`
	Icon        Image  `json:"icon"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type Image struct {
	Src   string            `json:"src"`
	Alt   string            `json:"alt"`
	Sizes map[string]string `json:"sizes,omitempty"`
}

type URLRef struct {
	DomainKey  string `json:"domainKey"`
	Path       string `json:"path"`
	ThemeColor string `json:"themeColor"`
}

type ProductVersion struct {
	Version  string                 `json:"version"`
	HostDocs bool                   `json:"hostDocs"`
	Show     bool                   `json:"show,omitempty"`
	DocsDir  string                 `json:"docsDir,omitempty"` // default: "docs"
	Branch   string                 `json:"branch,omitempty"`
	Info     map[string]interface{} `json:"info,omitempty"`
}

type Solution struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Image       Image  `json:"image"`
	Icon        Image  `json:"icon"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type Badge struct {
	URL  string `json:"url"`
	Alt  string `json:"alt"`
	Logo string `json:"logo"`
}

type Product struct {
	Key             string                `json:"key"`
	Name            string                `json:"name"`
	Fullname        string                `json:"fullname"`
	Tagline         string                `json:"tagline"`
	Summary         string                `json:"summary"`
	Published       bool                  `json:"published"`
	Author          string                `json:"author"`
	Website         *URLRef               `json:"website,omitempty"`
	Keywords        string                `json:"keywords,omitempty"`
	HeroImage       *Image                `json:"heroImage,omitempty"`
	HeroSlider      []Image               `json:"heroSlider,omitempty"`
	Logo            *Image                `json:"logo,omitempty"`
	LogoWhite       *Image                `json:"logoWhite,omitempty"`
	Icon            *Image                `json:"icon,omitempty"`
	RepoURL         string                `json:"repoURL,omitempty"`
	StarRepo        string                `json:"starRepo,omitempty"`
	DocRepo         string                `json:"docRepo,omitempty"`
	DatasheetFormID string                `json:"datasheetFormID,omitempty"`
	Badges          []Badge               `json:"badges,omitempty"`
	Videos          map[string]string     `json:"videos,omitempty"`
	Features        []Feature             `json:"features,omitempty"`
	Solutions       []Solution            `json:"solutions,omitempty"`
	Versions        []ProductVersion      `json:"versions,omitempty"`
	LatestVersion   string                `json:"latestVersion,omitempty"`
	SocialLinks     map[string]string     `json:"socialLinks,omitempty"`
	Description     map[string]string     `json:"description,omitempty"`
	SupportLinks    map[string]string     `json:"supportLinks,omitempty"`
	StripeProductID string                `json:"stripeProductID,omitempty"`
	Plans           map[string]Plan       `json:"plans,omitempty"`
	SubProjects     map[string]ProjectRef `json:"subProjects,omitempty"`
}

type Plan struct {
	Price json.Number `json:"price"`
}

type ProjectRef struct {
	Dir      string    `json:"dir"`
	Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
	Versions           []string `json:"versions"`
	SubProjectVersions []string `json:"subProjectVersions"`
}

type AssetListing struct {
	RepoURL string            `json:"repoURL"`
	Version string            `json:"version"`
	Dirs    map[string]string `json:"dirs"`
}

type Listing struct {
	Products []string     `json:"products"`
	Assets   AssetListing `json:"assets"`
}
