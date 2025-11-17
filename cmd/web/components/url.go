package components

func GetProductImageURL(path string) string {
	return "/cchoice/products/image?path=" + path
}

func GetBrandLogoURL(filename string) string {
	return "/cchoice/brands/logo?filename=" + filename
}

func GetAssetImageURL(filename string) string {
	return "/cchoice/assets/image?filename=" + filename
}
