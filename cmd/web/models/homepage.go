package models

type BrandImage struct {
	Filename     string
	Alt          string
	CDNURL       string
	IsComingSoon bool
}

type PostHomeContent struct {
	Title       string
	Description string
	BrandImage  *BrandImage
}

type PostHomeContentSection struct {
	Title  string
	Href   string
	Others []PostHomeContent
}

type HomePageData struct {
	Sections      []PostHomeContentSection
	StoreImageURL string
}

type BrandLogoURLFunc func(filename string) string

func BuildPostHomeContentSections(getBrandLogoURL BrandLogoURLFunc) []PostHomeContentSection {
	return []PostHomeContentSection{
		{
			Title: "About Us",
			Href:  "about-us",
			Others: []PostHomeContent{
				{
					Title:       "Who We Are",
					Description: "We are a dedicated company and trusted dealer, committed to selecting the right partners who provide high-quality tools and accessories to drive better progress in the future",
				},
				{
					Title:       "Vision",
					Description: "To be one of the leading importers of advance and innovative construction and fabrication tools and equipments",
				},
				{
					Title:       "Mission",
					Description: "To provide quality products certified by international manufacturers in both construction and fabrication materials industries",
				},
			},
		},
		{
			Title:  "Our Store",
			Href:   "store",
			Others: []PostHomeContent{},
		},
		{
			Title: "Our Services",
			Href:  "services",
			Others: []PostHomeContent{
				{
					Title:       "Sales",
					Description: "Quality construction tools and materials, competitively priced and sourced from trusted brands for every project scale",
				},
				{
					Title:       "Service",
					Description: "Reliable soonest delivery, after-sales support, product guidance, and technical assistance to keep your operations running smoothly",
				},
				{
					Title:       "Spare Parts",
					Description: "Genuine spare parts availability to extend equipment life and minimize downtime on site",
				},
			},
		},

		{
			Title: "Our Partners",
			Href:  "partners",
			Others: []PostHomeContent{
				{
					BrandImage: &BrandImage{
						Filename: "BOSCH.webp",
						Alt:      "Bosch image",
						CDNURL:   getBrandLogoURL("BOSCH.webp"),
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "TAILIN.webp",
						Alt:          "Tailin image",
						CDNURL:       getBrandLogoURL("TAILIN.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "BOSUN.webp",
						Alt:          "Bosun image",
						CDNURL:       getBrandLogoURL("BOSUN.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "NORTON.webp",
						Alt:          "NORTON image",
						CDNURL:       getBrandLogoURL("NORTON.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "DONGCHENG.webp",
						Alt:          "DONGCHENG image",
						CDNURL:       getBrandLogoURL("DONGCHENG.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "BERNMANN.webp",
						Alt:          "BERNMANN image",
						CDNURL:       getBrandLogoURL("BERNMANN.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "YOHINO.webp",
						Alt:          "YOHINO image",
						CDNURL:       getBrandLogoURL("YOHINO.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "ZEKOKI.webp",
						Alt:          "ZEKOKI image",
						CDNURL:       getBrandLogoURL("ZEKOKI.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "POWERCRAFT.webp",
						Alt:          "POWERCRAFT image",
						CDNURL:       getBrandLogoURL("POWERCRAFT.webp"),
						IsComingSoon: true,
					},
				},
				{
					BrandImage: &BrandImage{
						Filename:     "TATARA.webp",
						Alt:          "TATARA image",
						CDNURL:       getBrandLogoURL("TATARA.webp"),
						IsComingSoon: true,
					},
				},
			},
		},
	}
}
