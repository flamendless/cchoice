package requests

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/geocoding"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/VictoriaMetrics/fastcache"
	"golang.org/x/sync/singleflight"
)

func GetGeocodingCoordinates(
	cache *fastcache.Cache,
	sf *singleflight.Group,
	geocoder geocoding.IGeocoder,
	address string,
) (*geocoding.Coordinates, error) {
	cacheKey := []byte(generateGeocodingCacheKey(address))
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var coordinates geocoding.Coordinates
		if err := json.NewDecoder(buf).Decode(&coordinates); err == nil {
			return &coordinates, nil
		}
		metrics.Cache.MemHit()
	} else {
		metrics.Cache.MemMiss()
	}

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		coordinates, err := geocoder.GeocodeShippingAddress(address)
		if err != nil {
			return nil, err
		}
		return coordinates, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	coordinates := sfRes.(*geocoding.Coordinates)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(coordinates); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}
	return coordinates, nil
}

func generateGeocodingCacheKey(address string) string {
	cfg := conf.Conf()

	keyData := fmt.Sprintf("geocoding:%s:%s:%s",
		cfg.GeocodingService, cfg.GoogleMaps.APIKey[:8], address)

	hash := sha256.Sum256([]byte(keyData))
	return "geo_" + hex.EncodeToString(hash[:])[:16]
}
