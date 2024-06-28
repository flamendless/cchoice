package auth

import (
	"context"
	"errors"
)

type AuthToken struct {
	Token  string
	TSC bool
}

func ParseBearerToken(token string) (*AuthToken, error) {
	if token == "client" || token == "grpcui" {
		return &AuthToken{
			Token: token,
		}, nil
	}

	return nil, errors.New("Invalid token")
}

func UserClaimFromBearerToken(token *AuthToken) string {
	return token.Token
}

func (t AuthToken) GetRequestMetadata(
	ctx context.Context,
	in ...string,
) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.Token,
	}, nil
}

func (t AuthToken) RequireTransportSecurity() bool {
	return t.TSC
}
