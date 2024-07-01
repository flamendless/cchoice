package auth

//https://engineering.getweave.com/post/go-jwt-1/

import (
	"cchoice/cchoice_db"
	"cchoice/conf"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"context"
	"crypto"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	key crypto.PrivateKey
}

type Validator struct {
	key   crypto.PublicKey
	CtxDB *ctx.Database
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
	username string,
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
		"username": username,
	})

	tokenString, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	return tokenString, nil
}

func NewValidator(ctxDB *ctx.Database) (*Validator, error) {
	conf := conf.GetConf()
	key, err := jwt.ParseEdPublicKeyFromPEM(conf.PubKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse as ed private key: %w", err)
	}

	return &Validator{
		key:   key,
		CtxDB: ctxDB,
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

	claims := token.Claims.(jwt.MapClaims)
	aud, ok := claims["aud"]
	if !ok {
		return nil, fmt.Errorf("token had no audience claim")
	}

	audString := aud.(string)
	if enums.ParseAudEnum(audString) == enums.AudUndefined {
		return nil, fmt.Errorf("Token had the wrong audience claim: %s", aud)
	}

	_, err = v.CtxDB.QueriesRead.GetUserByEMailAndUserType(
		context.Background(),
		cchoice_db.GetUserByEMailAndUserTypeParams{
			Email:    claims["username"].(string),
			UserType: audString,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	return token, nil
}
