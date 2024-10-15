package auth

//https://engineering.getweave.com/post/go-jwt-1/

import (
	"cchoice/cchoice_db"
	"cchoice/conf"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/serialize"
	"context"
	"crypto"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ValidToken struct {
	Token       *jwt.Token
	TokenString string
	UserID      string
}

type Issuer struct {
	key    crypto.PrivateKey
	issuer string
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
		key:    key,
		issuer: conf.Issuer,
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
		"iss": i.issuer,

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

func (v *Validator) GetToken(
	expectedAUD enums.AudKind,
	tokenString string,
) (*ValidToken, error) {
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
		return nil, errors.New("token had no audience claim")
	}

	audString := aud.(string)
	eAUD := enums.ParseAudEnum(audString)
	if eAUD == enums.AUD_UNDEFINED || eAUD != expectedAUD {
		return nil, fmt.Errorf("Token had the wrong audience claim: %s", aud)
	}

	username := claims["username"].(string)

	var userID int64
	var errValidity error
	if expectedAUD == enums.AUD_API {
		userID, errValidity = v.CtxDB.Queries.GetUserByEMailAndUserTypeAndToken(
			context.Background(),
			cchoice_db.GetUserByEMailAndUserTypeAndTokenParams{
				Email:    username,
				UserType: audString,
				Token:    tokenString,
			},
		)
	} else if expectedAUD == enums.AUD_SYSTEM {
		userID, errValidity = v.CtxDB.Queries.GetUserByEMailAndUserType(
			context.Background(),
			cchoice_db.GetUserByEMailAndUserTypeParams{
				Email:    username,
				UserType: audString,
			},
		)
	}
	if errValidity != nil {
		return nil, fmt.Errorf("Invalid token: %w", errValidity)
	}

	return &ValidToken{
		TokenString: tokenString,
		Token:       token,
		UserID:      serialize.EncDBID(userID),
	}, nil
}
