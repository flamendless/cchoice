package conf

import (
	"cchoice/cmd/web/static"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"errors"
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

var appCfg *appConfig
var once sync.Once

type appConfig struct {
	Server             ServerConfig
	AppEnv             string `env:"APP_ENV" env-required:""`
	DBURL              string `env:"DB_URL" env-required:""`
	PaymentService     string `env:"PAYMENT_SERVICE" env-required:""`
	PayMongo           PayMongoConfig
	ShippingService    string `env:"SHIPPING_SERVICE" env-required:""`
	Lalamove           LalamoveConfig
	GeocodingService   string `env:"GEOCODING_SERVICE" env-required:""`
	GoogleMaps         GoogleMapsConfig
	OCRService         string `env:"OCR_SERVICE"`
	GoogleVisionConfig GoogleVisionConfig
	Business           BusinessConfig
	FSMode             string `env:"FSMODE" env-required:""`
	EncodeSalt         string `env:"ENCODE_SALT" env-required:""`
	LogMinLevel        int    `env:"LOG_MIN_LEVEL" env-default:"1"`
	StorageProvider    string `env:"STORAGE_PROVIDER" env-default:"local"`
	Linode             LinodeConfig
}

type ServerConfig struct {
	Address  string `env:"ADDRESS" env-required:""`
	Port     int    `env:"PORT" env-required:""`
	PortFS   int    `env:"PORT_FS" env-required:""`
	UseSSL   bool   `env:"USESSL"`
	UseHTTP2 bool   `env:"USEHTTP2"`
	CertPath string `env:"CERTPATH"`
	KeyPath  string `env:"KEYPATH"`
}

type PayMongoConfig struct {
	APIKey     string `env:"PAYMONGO_API_KEY"`
	BaseURL    string `env:"PAYMONGO_BASE_URL"`
	SuccessURL string `env:"PAYMONGO_SUCCESS_URL"`
	CancelURL  string `env:"PAYMONGO_CANCEL_URL"`
}

type LalamoveConfig struct {
	BaseURL string `env:"LALAMOVE_BASE_URL"`
	APIKey  string `env:"LALAMOVE_API_KEY"`
	Secret  string `env:"LALAMOVE_API_SECRET"`
}

type GoogleMapsConfig struct {
	APIKey string `env:"GOOGLE_MAPS_API_KEY"`
}

type GoogleVisionConfig struct {
	APIKey string `env:"GOOGLE_VISION_API_KEY"`
}

type BusinessConfig struct {
	Lat     string `env:"BUSINESS_LAT" env-default:"14.3866"`
	Lng     string `env:"BUSINESS_LNG" env-default:"120.8811"`
	Address string `env:"BUSINESS_ADDRESS" env-default:"General Trias, Cavite, Philippines"`
}

type LinodeConfig struct {
	Endpoint   string `env:"LINODE_ENDPOINT"`
	Region     string `env:"LINODE_REGION" env-default:""`
	BasePrefix string `env:"LINODE_BASE_PREFIX" env-default:""`

	Public  LinodeBucketConfig `env-prefix:"LINODE_PUBLIC_"`
	Private LinodeBucketConfig `env-prefix:"LINODE_PRIVATE_"`

	buckets map[enums.LinodeBucketEnum]LinodeBucketConfig
}

type LinodeBucketConfig struct {
	Bucket    string `env:"BUCKET"`
	AccessKey string `env:"ACCESS_KEY"`
	SecretKey string `env:"SECRET_KEY"`
}

func (lc *LinodeConfig) GetBuckets() map[enums.LinodeBucketEnum]LinodeBucketConfig {
	if lc.buckets == nil {
		lc.buckets = make(map[enums.LinodeBucketEnum]LinodeBucketConfig)
		lc.buckets[enums.LINODE_BUCKET_PUBLIC] = lc.Public
		lc.buckets[enums.LINODE_BUCKET_PRIVATE] = lc.Private
	}
	return lc.buckets
}

func (lc *LinodeConfig) GetBucketConfig(bucketEnum enums.LinodeBucketEnum) (LinodeBucketConfig, bool) {
	buckets := lc.GetBuckets()
	config, ok := buckets[bucketEnum]
	return config, ok
}

func mustValidate(c *appConfig) {
	if c.PaymentService == "paymongo" {
		if c.PayMongo.BaseURL == "" || c.PayMongo.APIKey == "" || c.PayMongo.SuccessURL == "" || c.PayMongo.CancelURL == "" {
			panic(errs.ErrPaymongoAPIKeyRequired)
		}
	} else {
		panic("Only 'paymongo' service is allowed for now")
	}

	switch c.ShippingService {
	case "lalamove":
		if c.Lalamove.BaseURL == "" || c.Lalamove.APIKey == "" || c.Lalamove.Secret == "" {
			panic(errs.ErrLalamoveAPIKeyRequired)
		}
	case "cchoice":
	default:
		panic("Only 'lalamove' or 'cchoice' service is allowed for now")
	}

	if c.GeocodingService == "googlemaps" {
		if c.GoogleMaps.APIKey == "" {
			panic(errs.ErrGMapsAPIKeyRequired)
		}
	} else {
		panic("Only 'googlemaps' service is allowed for now")
	}

	if c.OCRService == "googlevision" {
		if c.GoogleVisionConfig.APIKey == "" {
			panic(errs.ErrGVisionAPIKeyRequired)
		}
	}

	if c.IsLocal() {
		if c.Server.CertPath == "" || c.Server.KeyPath == "" {
			panic(fmt.Errorf("[CertPath, KeyPath]: %w", errs.ErrEnvVarRequired))
		}
	} else {
		c.Server.CertPath = fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", c.Server.Address)
		c.Server.KeyPath = fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", c.Server.Address)
	}

	if c.FSMode != static.GetMode() {
		panic(errors.Join(
			errs.ErrFS,
			fmt.Errorf(
				"got FSMODE '%s' but mode compiled was '%s'",
				c.FSMode,
				static.GetMode(),
			),
		))
	}

	if c.StorageProvider == "linode" {
		if c.Linode.Endpoint == "" || c.Linode.Region == "" {
			panic(fmt.Errorf("[Linode Storage]: %w", errs.ErrEnvVarRequired))
		}

		buckets := c.Linode.GetBuckets()
		for bucketEnum, bucketConfig := range buckets {
			if bucketConfig.Bucket == "" {
				continue
			}
			if bucketConfig.AccessKey == "" {
				panic(fmt.Errorf("[Linode Storage %s]: access key must be configure", bucketEnum.String()))
			}
			if bucketConfig.SecretKey == "" {
				panic(fmt.Errorf("[Linode Storage %s]: secret key must be configured", bucketEnum.String()))
			}
		}

		if len(buckets) != 2 {
			panic("[Linode Storage]: exactly two buckets must be configured")
		}
	}
}

func Conf() *appConfig {
	once.Do(func() {
		var co appConfig
		if err := cleanenv.ReadEnv(&co); err != nil {
			panic(errors.Join(errs.ErrEnvVarRequired, err))
		}
		mustValidate(&co)
		appCfg = &co
	})
	if appCfg == nil {
		panic("Conf should have been initialized at this point")
	}
	return appCfg
}

func (c *appConfig) IsLocal() bool {
	return c.AppEnv == "local"
}

func (c *appConfig) IsProd() bool {
	return c.AppEnv == "prod"
}
