package requests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sort"

	"cchoice/cmd/web/components/svg"
	"cchoice/cmd/web/models"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/a-h/templ"
	"golang.org/x/sync/singleflight"
)

type platformData struct {
	Icon  templ.Component
	Label string
	Order int
}

var platformKeys = []string{
	"url_main_shop",
	"url_facebook",
	// "url_shopee",
	"url_tiktok",
	// "url_youtube",
	"url_gmap",
	"url_waze",
}

var platformLabels = map[string]platformData{
	"url_main_shop": platformData{
		Label: "SHOP",
		Icon:  svg.Shop("size-10"),
		Order: 1,
	},
	"url_facebook": platformData{
		Label: "FACEBOOK",
		Icon:  svg.Facebook("size-10"),
		Order: 2,
	},
	"url_gmap": platformData{
		Label: "GOOGLE MAPS",
		Icon:  svg.GoogleMaps("size-10"),
		Order: 3,
	},
	"url_waze": platformData{
		Label: "WAZE",
		Icon:  svg.Waze("size-10"),
		Order: 4,
	},
	// "url_shopee": platformData{
	// 	Label: "SHOPEE",
	// 	Icon:  svg.Shopee("size-10"),
	// 	Order: 5,
	// },
	"url_tiktok": platformData{
		Label: "TIKTOK",
		Icon:  svg.TikTok("size-10"),
		Order: 5,
	},
	// "url_youtube": platformData{
	// 	Label: "YOUTUBE",
	// 	Icon:  svg.YouTube("size-10"),
	// 	Order: 5,
	// },
}

func init() {
	if len(platformKeys) != len(platformLabels) {
		panic("mismatch platform keys and labels length")
	}
}

func GetPlatforms(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	dbRO database.IService,
	cacheKey []byte,
) ([]models.Platform, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var res []models.Platform
		if err := gob.NewDecoder(buf).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	}
	metrics.Cache.MemMiss()

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		res, err := dbRO.GetQueries().GetSettingsByNames(ctx, platformKeys)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	settings := sfRes.([]queries.TblSetting)

	platforms := make([]models.Platform, 0, len(platformKeys))
	for _, setting := range settings {
		if setting.Value == "" {
			continue
		}
		p := platformLabels[setting.Name]
		platforms = append(platforms, models.Platform{
			Label: p.Label,
			Value: setting.Value,
			Icon:  p.Icon,
			Order: p.Order,
		})
	}

	sort.Slice(platforms, func(i, j int) bool {
		return platforms[i].Order < platforms[j].Order
	})

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(platforms); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return platforms, nil
}

func GeneratePlatformsCacheKey() []byte {
	keyData := "platforms:list"
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Appendf(nil, "plt_%s", hex.EncodeToString(hash[:])[:16])
}
