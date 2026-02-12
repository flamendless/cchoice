package requests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"cchoice/internal/utils"

	"github.com/VictoriaMetrics/fastcache"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

func GetSettingsData(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.Service,
	cacheKey []byte,
	keys []string,
) (map[string]string, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res map[string]string
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	} else {
		metrics.Cache.MemMiss()
	}

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetSettingsByNames(ctx, keys)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	res := sfRes.([]queries.TblSetting)

	settings := make(map[string]string, len(res))
	for _, setting := range res {
		settings[setting.Name] = setting.Value
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(settings); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return settings, nil
}

func GetCategoriesSidePanel(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.Service,
	cacheKey []byte,
) ([]models.CategorySidePanelText, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []models.CategorySidePanelText
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	} else {
		metrics.Cache.MemMiss()
	}

	const limit = 100
	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetProductCategoriesByPromoted(ctx, limit)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	res := sfRes.([]queries.GetProductCategoriesByPromotedRow)

	found := map[string]bool{}
	categories := make([]models.CategorySidePanelText, 0, len(res))
	for _, v := range res {
		label := utils.SlugToTile(v.Category.String)
		if _, exists := found[label]; exists {
			continue
		}
		categories = append(categories, models.CategorySidePanelText{
			Label:          label,
			URL:            "/product-category/" + v.Category.String,
			ScrollTargetID: utils.LabelToID(enums.MODULE_CATEGORY, label),
		})
		found[label] = true
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(categories); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return categories, nil
}

func GetBrandsSidePanel(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.Service,
	encoder encode.IEncode,
	cacheKey []byte,
) ([]models.BrandSidePanelText, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []models.BrandSidePanelText
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	} else {
		metrics.Cache.MemMiss()
	}

	const limit = 100
	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetBrandsForSidePanel(ctx, limit)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	res := sfRes.([]queries.GetBrandsForSidePanelRow)

	brands := make([]models.BrandSidePanelText, 0, len(res))
	for _, v := range res {
		brands = append(brands, models.BrandSidePanelText{
			Label:   v.Name,
			URL:     utils.URL("?brand=" + v.Name),
			BrandID: encoder.Encode(v.ID),
		})
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(brands); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return brands, nil
}

func GetCategorySectionHandler(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.Service,
	encoder encode.IEncode,
	cacheKey []byte,
	page int,
	limit int,
) ([]models.GroupedCategorySection, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []models.GroupedCategorySection
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	} else {
		metrics.Cache.MemMiss()
	}

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetProductCategoriesForSectionsPagination(
			ctx,
			queries.GetProductCategoriesForSectionsPaginationParams{
				Limit:  int64(limit),
				Offset: int64(page) * int64(limit),
			},
		)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)

	res := sfRes.([]queries.GetProductCategoriesForSectionsPaginationRow)
	categoriesSubcategories := map[string][]models.Subcategory{}
	for _, v := range res {
		if v.ProductsCount == 0 || v.Subcategory.String == "" {
			logs.Log().Debug(
				"category section has no prododuct or empty subcategory. Skipping...",
				zap.String("category name", v.Category.String),
				zap.String("subcategory name", v.Subcategory.String),
			)
			continue
		}

		category := utils.SlugToTile(v.Category.String)
		if _, exists := categoriesSubcategories[category]; !exists {
			categoriesSubcategories[category] = make([]models.Subcategory, 0, 8)
		}

		categoriesSubcategories[category] = append(categoriesSubcategories[category], models.Subcategory{
			CategoryID: encoder.Encode(v.ID),
			Label:      utils.SlugToTile(v.Subcategory.String),
		})
	}

	categorySections := make([]models.GroupedCategorySection, 0, len(categoriesSubcategories))
	for k, v := range categoriesSubcategories {
		sort.Slice(v, func(i, j int) bool {
			return v[i].Label < v[j].Label
		})
		categorySections = append(categorySections, models.GroupedCategorySection{
			Label:          k,
			ScrollTargetID: utils.LabelToID(enums.MODULE_CATEGORY, k),
			Subcategories:  v,
		})
	}
	sort.Slice(categorySections, func(i, j int) bool {
		return categorySections[i].Label < categorySections[j].Label
	})

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(categorySections); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return categorySections, nil
}

func GenerateSettingsCacheKey(keys []string) []byte {
	sort.Strings(keys)
	keyData := "homepage:settings:" + strings.Join(keys, ",")
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Appendf(nil, "hp_set_%s", hex.EncodeToString(hash[:])[:16])
}

func GenerateCategoriesSidePanelCacheKey(limit int64) []byte {
	keyData := fmt.Sprintf("homepage:categories_side:%d", limit)
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Appendf(nil, "hp_cat_%s", hex.EncodeToString(hash[:])[:16])
}

func GenerateCategorySectionCacheKey(page, limit int) []byte {
	keyData := fmt.Sprintf("homepage:category_sections:%d:%d", page, limit)
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Appendf(nil, "hp_sec_%s", hex.EncodeToString(hash[:])[:16])
}

func GetRandomSaleProduct(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.Service,
	encoder encode.IEncode,
	getCDNURL models.CDNURLFunc,
	cacheKey []byte,
) (*models.RandomSaleProduct, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res *models.RandomSaleProduct
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	} else {
		metrics.Cache.MemMiss()
	}

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetRandomProductOnSale(ctx)
		if err != nil {
			return nil, err
		}
		if res.ID == 0 {
			return nil, nil
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)

	if sfRes == nil {
		return nil, nil
	}

	product := sfRes.(queries.GetRandomProductOnSaleRow)
	origPrice, discountedPrice, discountPercentage := utils.GetOrigAndDiscounted(
		product.IsOnSale,
		product.UnitPriceWithVat,
		product.UnitPriceWithVatCurrency,
		product.SalePriceWithVat,
		product.SalePriceWithVatCurrency,
	)

	saleProduct := &models.RandomSaleProduct{
		GetRandomProductOnSaleRow: product,
		ProductID:                 encoder.Encode(product.ID),
		CDNURL:                    getCDNURL(product.ThumbnailPath),
		CDNURL1280:                getCDNURL(constants.ToPath1280(product.ThumbnailPath)),
		OrigPriceDisplay:          origPrice.Display(),
		PriceDisplay:              discountedPrice.Display(),
		DiscountPercentage:        discountPercentage,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(saleProduct); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return saleProduct, nil
}

func GenerateRandomSaleProductCacheKey(requestID string) []byte {
	keyData := "homepage:random_sale_product"
	if requestID == "" {
		requestID = utils.GenString(12)
	}
	keyData += ":" + requestID
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Appendf(nil, "hp_rsp_%s", hex.EncodeToString(hash[:])[:16])
}
