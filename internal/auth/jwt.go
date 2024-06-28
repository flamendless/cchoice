package auth

//https://engineering.getweave.com/post/go-jwt-1/

import (
	"cchoice/conf"
	"cchoice/internal/enums"
	"crypto"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	key crypto.PrivateKey
}

type Validator struct {
	key crypto.PublicKey
}

func NewIssuer() (*Issuer, error) {
	conf := conf.GetConf()
	key, err := jwt.ParseEdPrivateKeyFromPEM(conf.PrivKey)
	if err != nil {
		return nil, err
	}

	return &Issuer{
		key: key,
	}, nil
}

func (i *Issuer) IssueToken(
	aud enums.AudKind,
	user string,
) (string, error) {
	conf := conf.GetConf()
	now := time.Now()
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"aud": aud.String(),
		"nbf": now.Unix(),
		"iat": now.Unix(),
		"exp": now.Add(conf.TokenExp).Unix(),
		"iss": "http://localhost:8081",

		// other fields
		"user": user,
	})

	tokenString, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	return tokenString, nil
}

func NewValidator() (*Validator, error) {
	conf := conf.GetConf()
	key, err := jwt.ParseEdPublicKeyFromPEM(conf.PubKey)
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

	if aud != enums.AudAPI.String() && aud != enums.AudSystem.String() {
		return nil, fmt.Errorf("token had the wrong audience claim: %s", aud)
	}

	return token, nil
}
