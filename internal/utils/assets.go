package utils

//go:generate go run ../../cmd/genassets/genassets.go

var assetVersion = "dev"

func VersionedAsset(path string) string {
	return appendQueryParams(path, map[string]string{"v": assetVersion})
}
