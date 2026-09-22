package product

import "cchoice/cmd/web/models"

func heroImageURL(data models.ProductPageData) string {
	if data.Meta.OGImage != "" {
		return data.Meta.OGImage
	}
	if data.CDNURL != "" {
		return data.CDNURL
	}
	return data.CDNURL1280
}
