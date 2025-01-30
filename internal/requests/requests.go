package requests

import (
	"bytes"
	"context"
	"encoding/gob"

	"cchoice/cmd/web/models"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
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
		logs.Log().Debug("got value from cache")
		buf := bytes.NewBuffer(data)
		var res map[string]string
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			return nil, err
		}
		return res, nil
	}

	sfRes, err, shared := sf.Do(string(cacheKey), func() (interface{}, error) {
		res, err := dbRO.GetQueries().GetSettingsByNames(ctx, keys)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.Log().Debug("singleflight", zap.Bool("shared", shared))
	res := sfRes.([]queries.TblSetting)

	settings := make(map[string]string, len(res))
	for _, setting := range res {
		settings[setting.Name] = setting.Value
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(settings); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.Log().Debug("stored data to cache")
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
		logs.Log().Debug("got value from cache")
		buf := bytes.NewBuffer(data)
		var res []models.CategorySidePanelText
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			return nil, err
		}
		return res, nil
	}

	sfRes, err, shared := sf.Do(string(cacheKey), func() (interface{}, error) {
		res, err := dbRO.GetQueries().GetProductCategoriesByPromoted(ctx, params)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.Log().Debug("singleflight", zap.Bool("shared", shared))
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
		logs.Log().Debug("stored data to cache")
	}

	return categories, nil
}
