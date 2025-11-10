package components

func GetBrandLogoURL(s3URL string, filename string) string {
	if s3URL != "" {
		return s3URL
	}
	return "/cchoice/static/images/brand_logos/" + filename
}
