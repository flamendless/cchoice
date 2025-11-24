package requests

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/database"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"cchoice/internal/shipping"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/VictoriaMetrics/fastcache"
	"golang.org/x/sync/singleflight"
)

func GetShippingQuotation(
	ctx context.Context,
	cache *fastcache.Cache,
	sf *singleflight.Group,
	shippingService shipping.IShippingService,
	shippingRequest shipping.ShippingRequest,
	db database.Service,
) (*shipping.ShippingQuotation, error) {
	cacheKey := []byte(generateShippingCacheKey(
		shippingRequest.DeliveryLocation.Address,
		shippingRequest.Package.Weight,
		shippingRequest.ServiceType.String(),
	))

	if data, ok := cache.HasGet(nil, cacheKey); ok {
		buf := bytes.NewBuffer(data)
		var quotation shipping.ShippingQuotation
		if err := json.NewDecoder(buf).Decode(&quotation); err == nil {
			return &quotation, nil
		}
		metrics.Cache.MemHit()
	} else {
		metrics.Cache.MemMiss()
	}

	var quotation *shipping.ShippingQuotation

	sfRes, err, shared := sf.Do(string(cacheKey), func() (any, error) {
		q, e := shippingService.GetQuotation(shippingRequest)
		quotation = q

		logs.LogExternalAPICall(ctx, db.GetQueries(), logs.ExternalAPILogParams{
			CheckoutID: nil,
			Service:    "shipping",
			API:        shippingService.Enum(),
			Endpoint:   "/v3/quotations",
			HTTPMethod: "POST",
			Payload:    shippingRequest,
			Response:   q,
			Error:      e,
		})

		if e != nil {
			return nil, e
		}
		return q, nil
	})
	if err != nil {
		return nil, err
	}
	logs.SF(cacheKey, shared)
	quotation = sfRes.(*shipping.ShippingQuotation)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(quotation); err == nil {
		cache.Set(cacheKey, buf.Bytes())
		logs.CacheStore(cacheKey, buf)
	}
	return quotation, nil
}

func generateShippingCacheKey(address, weight, serviceType string) string {
	cfg := conf.Conf()

	keyData := fmt.Sprintf("shipping:%s:%s:%s:%s:%s:%s",
		cfg.Business.Lat, cfg.Business.Lng, address, weight, serviceType, cfg.ShippingService)

	hash := sha256.Sum256([]byte(keyData))
	return "ship_" + hex.EncodeToString(hash[:])[:16]
}
