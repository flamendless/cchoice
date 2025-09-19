package payments

import (
	"cchoice/internal/constants"
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
	PAYMENT_METHOD_COD

	// PAYMONGO
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

func (pm PaymentMethod) GetDisplayText() string {
	switch pm {
	case PAYMENT_METHOD_COD:
		return "Cash on Delivery"
	case PAYMENT_METHOD_QRPH:
		return "QRPH"
	case PAYMENT_METHOD_BILLEASE:
		return "Billease"
	case PAYMENT_METHOD_CARD:
		return "Card"
	case PAYMENT_METHOD_DOB:
		return "DOB"
	case PAYMENT_METHOD_DOB_UBP:
		return "Unionbank"
	case PAYMENT_METHOD_BRANKAS_BDO:
		return "BDO"
	case PAYMENT_METHOD_BRANKAS_LANDBANK:
		return "Landbank"
	case PAYMENT_METHOD_BRANKAS_METROBANK:
		return "Metrobank"
	case PAYMENT_METHOD_GCASH:
		return "GCash"
	case PAYMENT_METHOD_GRAB_PAY:
		return "GrabPay"
	case PAYMENT_METHOD_PAYMAYA:
		return "Maya"
	default:
		logs.Log().Warn("Unhandled payment display name", zap.Any("pm", pm))
		return pm.String()
	}
}

func (pm PaymentMethod) GetImageData(cache *fastcache.Cache, fs http.FileSystem) string {
	imgURL := constants.PathPaymentImages
	switch pm {
	case PAYMENT_METHOD_COD:
		imgURL += "cod.png"
	case PAYMENT_METHOD_QRPH:
		imgURL += "qrph.png"
	// case PAYMENT_METHOD_BILLEASE:
	// 	imgURL += "billease.svg"
	// case PAYMENT_METHOD_CARD:
	// 	imgURL += "qrph.png"
	// case PAYMENT_METHOD_DOB:
	// 	imgURL += "qrph.png"
	case PAYMENT_METHOD_DOB_UBP:
		imgURL += "unionbank.png"
	case PAYMENT_METHOD_BRANKAS_BDO:
		imgURL += "bdo.png"
	// case PAYMENT_METHOD_BRANKAS_LANDBANK:
	// 	imgURL += "qrph.png"
	// case PAYMENT_METHOD_BRANKAS_METROBANK:
	// 	imgURL += "qrph.png"
	case PAYMENT_METHOD_GCASH:
		imgURL += "gcash.png"
	// case PAYMENT_METHOD_GRAB_PAY:
	// 	imgURL += "qrph.png"
	case PAYMENT_METHOD_PAYMAYA:
		imgURL += "maya.png"
	default:
		logs.Log().Warn("Unhandled payment method", zap.Any("pm", pm))
		return ""
	}

	imgDataB64, err := images.GetImageDataB64(cache, fs, imgURL, images.IMAGE_FORMAT_PNG)
	if err != nil {
		logs.Log().Info("PaymentMethod image data", zap.Error(err))
		return ""
	}

	return imgDataB64
}

func ParsePaymentMethodToEnum(pm string) PaymentMethod {
	switch strings.ToUpper(pm) {
	case PAYMENT_METHOD_COD.String():
		return PAYMENT_METHOD_COD
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
