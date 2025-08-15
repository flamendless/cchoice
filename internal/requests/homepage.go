package requests

import (
	"bytes"
	"context"
	"encoding/gob"
	"sort"

	"cchoice/cmd/web/models"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
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
		logs.CacheHit(cacheKey, len(res))
		return res, nil
	} else {
		logs.CacheMiss(cacheKey)
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
	params queries.GetProductCategoriesByPromotedParams,
) ([]models.CategorySidePanelText, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []models.CategorySidePanelText
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		logs.CacheHit(cacheKey, len(res))
		return res, nil
	} else {
		logs.CacheMiss(cacheKey)
	}

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetProductCategoriesByPromoted(ctx, params)
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
			Label: label,
			URL:   "/product-category/" + v.Category.String,
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
		logs.CacheHit(cacheKey, len(res))
		return res, nil
	} else {
		logs.CacheMiss(cacheKey)
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
			Label:         k,
			Subcategories: v,
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
