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
	StorageProvider    string `env:"STORAGE_PROVIDER" env-default:"LOCAL"`
	Linode             LinodeConfig
	CloudflareImages   CloudflareImagesConfig
	MailService        string `env:"MAIL_SERVICE"`
	MailerooConfig     MailerooConfig
	Settings           Settings
}

type Settings struct {
	MobileNo    string
	EMail       string
	Address     string
	URLGMap     string
	URLFacebook string
	URLTikTok   string
}

type MailerooConfig struct {
	APIKey string `env:"MAILEROO_API_KEY"`
	From   string `env:"MAILEROO_FROM"`
	CC     string `env:"MAILEROO_CC"`
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
	APIKey           string `env:"PAYMONGO_API_KEY"`
	BaseURL          string `env:"PAYMONGO_BASE_URL"`
	SuccessURL       string `env:"PAYMONGO_SUCCESS_URL"`
	CancelURL        string `env:"PAYMONGO_CANCEL_URL"`
	WebhookSecretKey string `env:"PAYMONGO_WEBHOOK_SECRET_KEY"`
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
	Lat        string `env:"BUSINESS_LAT" env-required:""`
	Lng        string `env:"BUSINESS_LNG" env-required:""`
	Address    string `env:"BUSINESS_ADDRESS" env-required:""`
	Line1      string `env:"BUSINESS_LINE1" env-required:""`
	Line2      string `env:"BUSINESS_LINE2" env-required:""`
	City       string `env:"BUSINESS_CITY" env-required:""`
	State      string `env:"BUSINESS_STATE" env-required:""`
	PostalCode string `env:"BUSINESS_POSTAL_CODE" env-required:""`
	Country    string `env:"BUSINESS_COUNTRY" env-required:""`
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

type CloudflareImagesConfig struct {
	AccountID   string `env:"CLOUDFLARE_ACCOUNT_ID"`
	AccountHash string `env:"CLOUDFLARE_ACCOUNT_HASH"`
	APIToken    string `env:"CLOUDFLARE_IMAGES_API_TOKEN"`
	Variant     string `env:"CLOUDFLARE_IMAGES_VARIANT" env-default:"public"`
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

	if c.Business.Lat == "" || c.Business.Lng == "" || c.Business.Address == "" || c.Business.Line1 == "" || c.Business.Line2 == "" || c.Business.City == "" || c.Business.State == "" || c.Business.PostalCode == "" || c.Business.Country == "" {
		panic(fmt.Errorf("[Business]: %w", errs.ErrEnvVarRequired))
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

func (c *appConfig) SetSettings(settings map[string]string) {
	c.Settings.MobileNo = settings["mobile_no"]
	c.Settings.EMail = settings["email"]
	c.Settings.Address = settings["address"]
	c.Settings.URLGMap = settings["url_gmap"]
	c.Settings.URLFacebook = settings["url_facebook"]
	c.Settings.URLTikTok = settings["url_tiktok"]
}

func (c *appConfig) IsLocal() bool {
	return c.AppEnv == "local"
}

func (c *appConfig) IsProd() bool {
	return c.AppEnv == "prod"
}

func (c *appConfig) IsWeb() bool {
	return c.AppEnv == "web"
}
