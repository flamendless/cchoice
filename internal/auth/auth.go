package auth

import (
	"context"
)

type AuthToken struct {
	Token string
	Sub   string
	TSC   bool
}

func ParseToken(token string) (*AuthToken, error) {
	return &AuthToken{
		Token: token,
	}, nil
}

func UserClaimFromToken(token *AuthToken) string {
	return token.Sub
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
