package models

type BrandImage struct {
	Filename string
	Alt      string
	CDNURL   string
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
					Title:       "Delivery",
					Description: "Expect delivery at soonest",
				},
				{
					Title:       "Here when you need us",
					Description: "Always available and right around the corner",
				},
				{
					Title:       "Expect high-quality products",
					Description: "Authorized and trusted",
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
						Filename: "TAILIN.webp",
						Alt:      "Tailin image",
						CDNURL:   getBrandLogoURL("TAILIN.webp"),
					},
				},
				{
					BrandImage: &BrandImage{
						Filename: "DELTAPLUS.webp",
						Alt:      "DeltaPlus image",
						CDNURL:   getBrandLogoURL("DELTAPLUS.webp"),
					},
				},
				{
					BrandImage: &BrandImage{
						Filename: "BOSUN.webp",
						Alt:      "Bosun image",
						CDNURL:   getBrandLogoURL("BOSUN.webp"),
					},
				},
				{
					BrandImage: &BrandImage{
						Filename: "STANLEY.webp",
						Alt:      "Stanley image",
						CDNURL:   getBrandLogoURL("STANLEY.webp"),
					},
				},
			},
		},
	}
}
