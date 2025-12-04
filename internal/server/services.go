package server

import (
	"cchoice/internal/conf"
	"cchoice/internal/database"
	"cchoice/internal/enums"
	"cchoice/internal/geocoding"
	"cchoice/internal/geocoding/googlemaps"
	"cchoice/internal/mail"
	"cchoice/internal/mail/maileroo"
	"cchoice/internal/payments"
	"cchoice/internal/payments/paymongo"
	"cchoice/internal/shipping"
	cchoiceservice "cchoice/internal/shipping/cchoice"
	"cchoice/internal/shipping/lalamove"
	"cchoice/internal/storage"
	"cchoice/internal/storage/linode"
	localstorage "cchoice/internal/storage/local"
)

func mustInitPaymentGateway() payments.IPaymentGateway {
	cfg := conf.Conf()
	switch cfg.PaymentService {
	case payments.PAYMENT_GATEWAY_PAYMONGO.String():
		return paymongo.MustInit()
	default:
		panic("Unsupported payment service: " + cfg.PaymentService)
	}
}

func mustInitShippingService() shipping.IShippingService {
	cfg := conf.Conf()
	switch cfg.ShippingService {
	case shipping.SHIPPING_SERVICE_LALAMOVE.String():
		return lalamove.MustInit()
	case shipping.SHIPPING_SERVICE_CCHOICE.String():
		return cchoiceservice.MustInit()
	default:
		panic("Unsupported shipping service: " + cfg.ShippingService)
	}
}

func mustInitGeocodingService(dbRW database.Service) geocoding.IGeocoder {
	cfg := conf.Conf()
	switch cfg.GeocodingService {
	case geocoding.GEOCODING_SERVICE_GOOGLEMAPS.String():
		return googlemaps.MustInit(dbRW)
	default:
		panic("Unsupported geocoding service: " + cfg.GeocodingService)
	}
}

func mustInitStorageProvider() (storage.IObjectStorage, storage.IFileSystem) {
	cfg := conf.Conf()
	switch cfg.StorageProvider {
	case storage.STORAGE_PROVIDER_LOCAL.String():
		return nil, localstorage.New()
	case storage.STORAGE_PROVIDER_LINODE.String():
		objStorage := linode.MustInitWithBucket(enums.LINODE_BUCKET_PUBLIC)

		//TODO: (Brandon) product images should be in private bucket
		productImageFS := linode.New(objStorage)
		return objStorage, productImageFS
	default:
		panic("Unsupported storage provider: " + cfg.StorageProvider)
	}
}

func mustInitMailService() mail.IMailService {
	cfg := conf.Conf()
	switch cfg.MailService {
	case mail.MAIL_SERVICE_MAILEROO.String():
		return maileroo.MustInit()
	default:
		panic("Unsupported mail service: " + cfg.MailService)
	}
}
