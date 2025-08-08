package payments

import (
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"fmt"
	"net/http"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

//go:generate go tool stringer -type=PaymentMethod -trimprefix=PAYMENT_METHOD_

type PaymentMethod int

const (
	PAYMENT_METHOD_UNDEFINED PaymentMethod = iota
	PAYMENT_METHOD_QRPH
	PAYMENT_METHOD_BILLEASE
	PAYMENT_METHOD_CARD
	PAYMENT_METHOD_DOB
	PAYMENT_METHOD_DOB_UBP
	PAYMENT_METHOD_BRANKAS_BDO
	PAYMENT_METHOD_BRANKAS_LANDBANK
	PAYMENT_METHOD_BRANKAS_METROBANK
	PAYMENT_METHOD_GCASH
	PAYMENT_METHOD_GRAB_PAY
	PAYMENT_METHOD_PAYMAYA
)

func (pm PaymentMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToLower(pm.String()))
}

func (pm PaymentMethod) GetImageData(cache *fastcache.Cache, fs http.FileSystem) string {
	var imgURL string
	switch pm {
	case PAYMENT_METHOD_QRPH:
		imgURL = "static/images/payments/qrph.png"
	default:
		panic("Unhandled payment method image")
	}

	finalPath, ext, err := images.GetImagePathWithSize(imgURL, "", false)
	if err != nil {
		logs.Log().Info("PaymentMethod image data", zap.Error(err))
		return ""
	}
	imgDataB64, err := images.GetImageDataB64(cache, fs, finalPath, ext)
	if err != nil {
		logs.Log().Info("PaymentMethod image data", zap.Error(err))
		return ""
	}
	return imgDataB64
}

func ParsePaymentMethodToEnum(pm string) PaymentMethod {
	switch strings.ToUpper(pm) {
	case PAYMENT_METHOD_QRPH.String():
		return PAYMENT_METHOD_QRPH
	case PAYMENT_METHOD_BILLEASE.String():
		return PAYMENT_METHOD_BILLEASE
	case PAYMENT_METHOD_CARD.String():
		return PAYMENT_METHOD_CARD
	case PAYMENT_METHOD_DOB.String():
		return PAYMENT_METHOD_DOB
	case PAYMENT_METHOD_DOB_UBP.String():
		return PAYMENT_METHOD_DOB_UBP
	case PAYMENT_METHOD_BRANKAS_BDO.String():
		return PAYMENT_METHOD_BRANKAS_BDO
	case PAYMENT_METHOD_BRANKAS_LANDBANK.String():
		return PAYMENT_METHOD_BRANKAS_LANDBANK
	case PAYMENT_METHOD_BRANKAS_METROBANK.String():
		return PAYMENT_METHOD_BRANKAS_METROBANK
	case PAYMENT_METHOD_GCASH.String():
		return PAYMENT_METHOD_GCASH
	case PAYMENT_METHOD_GRAB_PAY.String():
		return PAYMENT_METHOD_GRAB_PAY
	case PAYMENT_METHOD_PAYMAYA.String():
		return PAYMENT_METHOD_PAYMAYA
	default:
		panic(fmt.Errorf("undefined payment method '%s'", pm))
	}
}
