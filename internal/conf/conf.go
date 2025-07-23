package conf

import (
	"cchoice/internal/errs"
	"errors"
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

var appCfg *appConfig
var once sync.Once

type appConfig struct {
	Address            string `env:"ADDRESS" env-required:""`
	Port               int    `env:"PORT" env-required:""`
	PortFS             int    `env:"PORT_FS" env-required:""`
	AppEnv             string `env:"APP_ENV" env-required:""`
	DBURL              string `env:"DB_URL" env-required:""`
	PaymentService     string `env:"PAYMENT_SERVICE" env-required:""`
	PayMongoAPIKey     string `env:"PAYMONGO_API_KEY"`
	PayMongoSuccessURL string `env:"PAYMONGO_SUCCESS_URL"`
	EncodeSalt         string `env:"ENCODE_SALT" env-required:""`
	LogMinLevel        int    `env:"LOG_MIN_LEVEL" env-default:"1"`
	UseSSL             bool   `env:"USESSL"`
	UseHTTP2           bool   `env:"USEHTTP2"`
	CertPath           string `env:"CERTPATH"`
	KeyPath            string `env:"KEYPATH"`
}

func mustValidate(c *appConfig) {
	if c.PaymentService == "paymongo" {
		if c.PayMongoAPIKey == "" || c.PayMongoSuccessURL == "" {
			panic(fmt.Errorf("[PayMongo]: %w", errs.ErrEnvVarRequired))
		}
	} else {
		panic("Only 'paymongo' service is allowed for now")
	}

	if c.IsLocal() {
		if c.CertPath == "" || c.KeyPath == "" {
			panic(fmt.Errorf("[CertPath, KeyPath]: %w", errs.ErrEnvVarRequired))
		}
	} else {
		c.CertPath = fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", c.Address)
		c.KeyPath = fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", c.Address)
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
	return appCfg
}

func (c *appConfig) IsLocal() bool {
	return c.AppEnv == "local"
}

func (c *appConfig) IsProd() bool {
	return c.AppEnv == "prod"
}
