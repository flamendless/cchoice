package seo

import (
	"encoding/json"
	"strings"
)

const (
	SiteTitle = "C-Choice | Power Tools & Construction Supplies Philippines"

	SiteDescription = "Shop power tools, construction tools, and construction supplies in the Philippines. " +
		"C-Choice (cchoice) carries Bosch, INGCO, and trusted brands for contractors and builders. Your partner in progress."

	SiteKeywords = "cchoice, c-choice, power tools, construction tools, construction supply, construction supplies, " +
		"philippines, bosch philippines, ingco philippines, bosch, ingco, bosh, hardware store philippines, " +
		"building materials philippines, tools philippines"

	SiteOGTitle       = "C-Choice | Power Tools & Construction Supplies Philippines"
	SiteOGDescription = "Shop Bosch, INGCO, and quality power tools and construction supplies online in the Philippines."
)

func BaseProductKeywords() []string {
	return []string{
		"cchoice",
		"c-choice",
		"power tools",
		"construction tools",
		"construction supplies",
		"philippines",
	}
}

func BuildSiteStructuredData(homeURL, logoURL string) json.RawMessage {
	homeURL = strings.TrimSuffix(homeURL, "/")

	type organization struct {
		Context       string   `json:"@context"`
		Type          string   `json:"@type"`
		Name          string   `json:"name"`
		AlternateName []string `json:"alternateName"`
		URL           string   `json:"url"`
		Logo          string   `json:"logo"`
		Description   string   `json:"description"`
		AreaServed    string   `json:"areaServed"`
		KnowsAbout    []string `json:"knowsAbout"`
	}
	type webSite struct {
		Context string `json:"@context"`
		Type    string `json:"@type"`
		Name    string `json:"name"`
		URL     string `json:"url"`
	}
	type graph struct {
		Context string `json:"@context"`
		Graph   []any  `json:"@graph"`
	}

	if logoURL == "" {
		logoURL = DefaultOGImage
	}

	payload := graph{
		Context: "https://schema.org",
		Graph: []any{
			organization{
				Context: "https://schema.org",
				Type:    "Organization",
				Name:    "C-Choice Construction Supply",
				AlternateName: []string{
					"C-Choice",
					"CChoice",
					"cchoice",
				},
				URL:         homeURL + "/",
				Logo:        logoURL,
				Description: SiteDescription,
				AreaServed:  "Philippines",
				KnowsAbout: []string{
					"Power Tools",
					"Construction Tools",
					"Construction Supplies",
					"Bosch Philippines",
					"INGCO Philippines",
					"Bosch",
					"INGCO",
				},
			},
			webSite{
				Context: "https://schema.org",
				Type:    "WebSite",
				Name:    "C-Choice Construction Supply",
				URL:     homeURL + "/",
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return data
}
