package requests

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"

	"cchoice/cmd/parse_map/enums"
	"cchoice/cmd/parse_map/models"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"

	"github.com/VictoriaMetrics/fastcache"
	"golang.org/x/sync/singleflight"
)

func GetProvinces(
	cache *fastcache.Cache,
	sf *singleflight.Group,
) ([]*models.Map, error) {
	cacheKey := []byte(generateProvincesCacheKey())
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []*models.Map
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
		provinces := make([]*models.Map, 0)
		for _, region := range models.PhilippinesMap {
			for _, province := range region.Contents {
				if province.Level == enums.LEVEL_PROVINCE {
					provinceCopy := *province
					provinceCopy.Contents = nil
					provinces = append(provinces, &provinceCopy)
				}
			}
		}
		return provinces, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	maps := sfRes.([]*models.Map)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(maps); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}
	return maps, nil
}

func GetCitiesByProvince(
	cache *fastcache.Cache,
	sf *singleflight.Group,
	province string,
) ([]*models.Map, error) {
	if province == "" {
		return nil, errs.ErrInvalidParams
	}

	cacheKey := []byte(generateCitiesCacheKey(province))

	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []*models.Map
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
		provinceMap := models.BinarySearchMapByName(models.PhilippinesMap, province, enums.LEVEL_PROVINCE)
		if provinceMap == nil {
			return nil, errs.ErrInvalidParams
		}
		return provinceMap.Contents, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	maps := sfRes.([]*models.Map)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(maps); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}
	return maps, nil
}

func GetBarangaysByCity(
	cache *fastcache.Cache,
	sf *singleflight.Group,
	city string,
) ([]*models.Map, error) {
	if city == "" {
		return nil, errs.ErrInvalidParams
	}

	cacheKey := []byte(generateBarangaysCacheKey(city))

	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []*models.Map
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
		cityMap := models.BinarySearchMapByName(models.PhilippinesMap, city, enums.LEVEL_CITY)
		if cityMap == nil {
			cityMap = models.BinarySearchMapByName(models.PhilippinesMap, city, enums.LEVEL_MUNICIPALITY)
			if cityMap == nil {
				return nil, errs.ErrInvalidParams
			}
			return cityMap.Contents, nil
		}
		return cityMap.Contents, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	maps := sfRes.([]*models.Map)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(maps); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}
	return maps, nil
}

func generateProvincesCacheKey() string {
	keyData := "maps:provinces"
	hash := sha256.Sum256([]byte(keyData))
	return "map_prov_" + hex.EncodeToString(hash[:])[:16]
}

func generateCitiesCacheKey(province string) string {
	keyData := "maps:cities:" + province
	hash := sha256.Sum256([]byte(keyData))
	return "map_city_" + hex.EncodeToString(hash[:])[:16]
}

func generateBarangaysCacheKey(city string) string {
	keyData := "maps:barangays:" + city
	hash := sha256.Sum256([]byte(keyData))
	return "map_brgy_" + hex.EncodeToString(hash[:])[:16]
}
