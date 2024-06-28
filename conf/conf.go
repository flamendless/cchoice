package conf

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type config struct {
	PrivKey        string        `env:"PRIVKEY,required"`
	PubKey         string        `env:"PUBKEY,required"`
	TokenExp       time.Duration `env:"TokenExp"`
	ClientUsername string        `env:"ClientUsername,required"`
}

type Config struct {
	PrivKey        []byte
	PubKey         []byte
	TokenExp       time.Duration
	ClientUsername string
}

var conf Config

func init() {
	LoadConf()
}

func LoadConf() {
	configMd := config{
		TokenExp: time.Minute * 10,
	}
	err := env.Parse(&configMd)
	if err != nil {
		panic(err)
	}

	conf = Config{
		PrivKey:        []byte(configMd.PrivKey),
		PubKey:         []byte(configMd.PubKey),
		TokenExp:       configMd.TokenExp,
		ClientUsername: configMd.ClientUsername,
	}
}

func GetConf() Config {
	return conf
}
