package server

import (
	"sync"

	"cchoice/internal/storage"
)

var cdnURLCache sync.Map

func (s *Server) GetCDNURL(path string) string {
	if cached, ok := cdnURLCache.Load(path); ok {
		return cached.(string)
	}

	var url string
	if s.objectStorage != nil && s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES {
		url = s.objectStorage.GetPublicURL(path)
	}
	if url == "" {
		url = "/cchoice/products/image?thumbnail=1&quality=best&path=" + path
	}

	cdnURLCache.Store(path, url)
	return url
}

func (s *Server) GetBrandLogoCDNURL(filename string) string {
	cacheKey := "brand:" + filename
	if cached, ok := cdnURLCache.Load(cacheKey); ok {
		return cached.(string)
	}

	var url string
	if s.objectStorage != nil && s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES {
		path := "static/images/brand_logos/" + filename
		url = s.objectStorage.GetPublicURL(path)
	}
	if url == "" {
		url = "/cchoice/brands/logo?filename=" + filename
	}

	cdnURLCache.Store(cacheKey, url)
	return url
}

func (s *Server) GetAssetCDNURL(filename string) string {
	cacheKey := "asset:" + filename
	if cached, ok := cdnURLCache.Load(cacheKey); ok {
		return cached.(string)
	}

	var url string
	if s.objectStorage != nil && s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES {
		path := "static/images/" + filename
		url = s.objectStorage.GetPublicURL(path)
	}
	if url == "" {
		url = "/cchoice/assets/image?filename=" + filename
	}

	cdnURLCache.Store(cacheKey, url)
	return url
}
