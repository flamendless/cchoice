package middlewares

import (
	"context"

	cchoiceauth "cchoice/internal/auth"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var CtxToken cchoiceauth.AuthToken

func AuthBearer(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	bearerToken, err := cchoiceauth.ParseBearerToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid auth token: %v", err)
	}

	ctx = logging.InjectFields(
		ctx,
		logging.Fields{"auth.sub", cchoiceauth.UserClaimFromBearerToken(bearerToken)},
	)

	return context.WithValue(ctx, CtxToken, bearerToken), nil
}
