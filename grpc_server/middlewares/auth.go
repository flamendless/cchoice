package middlewares

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BearerToken struct {
	Token string
}

var CtxToken BearerToken

func parseToken(token string) (BearerToken, error) {
	return BearerToken{
		Token: token,
	}, nil
}

func userClaimFromToken(token *BearerToken) string {
	return token.Token
}

func AuthBearer(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	tokenInfo, err := parseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid auth token: %v", err)
	}

	ctx = logging.InjectFields(
		ctx,
		logging.Fields{"auth.sub", userClaimFromToken(&tokenInfo)},
	)

	return context.WithValue(ctx, CtxToken, tokenInfo), nil
}
