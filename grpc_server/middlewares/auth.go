package middlewares

import (
	"context"

	cchoiceauth "cchoice/internal/auth"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var CtxToken cchoiceauth.ClientToken

type AuthMiddleware struct {
	Validator *cchoiceauth.Validator
}

func AddAuth(validator *cchoiceauth.Validator) AuthMiddleware {
	return AuthMiddleware{
		Validator: validator,
	}
}

func (mw *AuthMiddleware) Handle(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	jwtToken, err := mw.Validator.GetToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid auth token: %v", err)
	}

	// ctx = logging.InjectFields(
	// 	ctx,
	// 	logging.Fields{"auth.sub", cchoiceauth.UserClaimFromToken(bearerToken)},
	// )

	return context.WithValue(ctx, CtxToken, jwtToken), nil
}
