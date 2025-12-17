package requests

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"

	"github.com/VictoriaMetrics/fastcache"
	"golang.org/x/sync/singleflight"

	"cchoice/internal/changelogs"
	"cchoice/internal/constants"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
)

func GetChangeLogs(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	cacheKey []byte,
	appenv string,
	limit int,
) ([]changelogs.ChangeLog, error) {
	if data, ok := cache.HasGet(nil, cacheKey); ok {
		var res []changelogs.ChangeLog
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&res); err != nil {
			logs.GobError(cacheKey, err)
			return nil, err
		}
		metrics.Cache.MemHit()
		return res, nil
	}
	metrics.Cache.MemMiss()

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		f, err := os.Open(constants.PathChangelogs)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return changelogs.Parse(f, appenv, limit)
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)

	res := sfRes.([]changelogs.ChangeLog)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(res); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}

	return res, nil
}
