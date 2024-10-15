package conf

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type config struct {
	PrivKey        string        `env:"PRIVKEY,required"`
	PubKey         string        `env:"PUBKEY,required"`
	Mode           string        `env:"MODE"`
	TokenExp       time.Duration `env:"TOKENEXP"`
	ClientUsername string        `env:"CLIENTUSERNAME,required"`
	Issuer         string        `env:"ISSUER,required"`
}

type Config struct {
	PrivKey        []byte
	PubKey         []byte
	Mode           string
	TokenExp       time.Duration
	ClientUsername string
	Issuer         string
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
		Mode:           configMd.Mode,
		TokenExp:       configMd.TokenExp,
		ClientUsername: configMd.ClientUsername,
		Issuer:         configMd.Issuer,
	}
}

func GetConf() Config {
	return conf
}
