package auth

//https://engineering.getweave.com/post/go-jwt-1/

import (
	"crypto"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	PrivKey string `env:"PRIVKEY,required"`
	PubKey  string `env:"PUBKEY,required"`
}

type Issuer struct {
	key crypto.PrivateKey
}

type Validator struct {
	key crypto.PublicKey
}

var (
	jwtConf JWTConfig
	privkey []byte
	pubkey  []byte
)

func init() {
	jwtConf := JWTConfig{}
	err := env.Parse(&jwtConf)
	if err != nil {
		panic(err)
	}
	privkey = []byte(jwtConf.PrivKey)
	pubkey = []byte(jwtConf.PubKey)
}

func NewIssuer() (*Issuer, error) {
	key, err := jwt.ParseEdPrivateKeyFromPEM(privkey)
	if err != nil {
		return nil, err
	}

	return &Issuer{
		key: key,
	}, nil
}

func (i *Issuer) IssueToken(user string, roles []string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"aud": "api",
		"nbf": now.Unix(),
		"iat": now.Unix(),
		"exp": now.Add(time.Minute).Unix(),
		"iss": "http://localhost:8081",

		// other fields
		"user":  user,
		"roles": roles,
	})

	tokenString, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	return tokenString, nil
}

func NewValidator() (*Validator, error) {
	key, err := jwt.ParseEdPublicKeyFromPEM(pubkey)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse as ed private key: %w", err)
	}

	return &Validator{
		key: key,
	}, nil
}

func (v *Validator) GetToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodEd25519)
			if !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return v.key, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Unable to parse token string: %w", err)
	}

	aud, ok := token.Claims.(jwt.MapClaims)["aud"]
	if !ok {
		return nil, fmt.Errorf("token had no audience claim")
	}

	if aud != "api" && aud != "system" {
		return nil, fmt.Errorf("token had the wrong audience claim")
	}

	return token, nil
}
