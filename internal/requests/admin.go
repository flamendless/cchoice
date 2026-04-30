package requests

import (
	"bytes"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"slices"

	"github.com/VictoriaMetrics/fastcache"
	"golang.org/x/sync/singleflight"
)

func GenerateAdminBrandsCacheKey() []byte {
	keyData := "admin:brands"
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Appendf(nil, "adm_br_%s", hex.EncodeToString(hash[:])[:16])
}

func GenerateAdminCategoriesCacheKey() []byte {
	keyData := "admin:categories"
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Appendf(nil, "adm_cat_%s", hex.EncodeToString(hash[:])[:16])
}

func GetBrandsForAdmin(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.IService,
	cacheKey []byte,
) ([]queries.GetBrandsForProductCreateRow, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []queries.GetBrandsForProductCreateRow
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	} else {
		metrics.Cache.MemMiss()
	}

	const limit = 500
	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetBrandsForProductCreate(ctx, limit)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	res := sfRes.([]queries.GetBrandsForProductCreateRow)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(res); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return res, nil
}

func GetCategoriesForAdmin(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.IService,
	cacheKey []byte,
) (map[string][]string, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res map[string][]string
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
		res, err := dbRO.GetQueries().GetProductCategoriesForSections(ctx)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	dbRes := sfRes.([]queries.GetProductCategoriesForSectionsRow)

	categories := make(map[string][]string)
	for _, v := range dbRes {
		if !v.Category.Valid || v.Category.String == "" {
			continue
		}
		cat := v.Category.String
		subcat := ""
		if v.Subcategory.Valid {
			subcat = v.Subcategory.String
		}
		if _, exists := categories[cat]; !exists {
			categories[cat] = []string{}
		}
		if subcat != "" {
			hasSubcat := slices.Contains(categories[cat], subcat)
			if !hasSubcat {
				categories[cat] = append(categories[cat], subcat)
			}
		}
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(categories); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return categories, nil
}
