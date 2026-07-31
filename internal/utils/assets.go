package utils

import "cchoice/internal/constants"

//go:generate go run ../../cmd/genassets/genassets.go

var assetVersion = "dev"

func VersionedAsset(path string) string {
	return appendQueryParams(path, map[string]string{constants.QueryParamAssetVersion: assetVersion})
}
