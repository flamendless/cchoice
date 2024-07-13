package middlewares

import (
	"cchoice/client/common"
	"cchoice/internal/auth"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	pb "cchoice/proto"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type Authenticated struct {
	AuthService pb.AuthServiceClient
	SM          *scs.SessionManager
}

func NewAuthenticated(
	sm *scs.SessionManager,
	authService pb.AuthServiceClient,
) *Authenticated {
	return &Authenticated{
		SM:          sm,
		AuthService: authService,
	}
}

func (mw Authenticated) Authenticated(
	w http.ResponseWriter,
	r *http.Request,
	aud enums.AudKind,
) (*auth.ValidToken, error) {
	authSession, ok := mw.SM.Get(r.Context(), "authSession").(common.AuthSession)
	if !ok || authSession.Token == "" {
		return nil, errs.ERR_EXPIRED_SESSION
	}
	if authSession.NeedOTP {
		return nil, errs.ERR_NEED_OTP
	}

	res, err := mw.AuthService.ValidateToken(
		r.Context(),
		&pb.ValidateTokenRequest{
			Token: authSession.Token,
			Aud:   aud.String(),
		},
	)
	if err != nil {
		return nil, errs.ERR_INVALID_TOKEN
	}

	return &auth.ValidToken{
		UserID:      res.UserId,
		TokenString: res.TokenString,
	}, nil
}

func (mw Authenticated) AuthenticatedSkipOTP(
	w http.ResponseWriter,
	r *http.Request,
	aud enums.AudKind,
) (*auth.ValidToken, error) {
	authSession, ok := mw.SM.Get(r.Context(), "authSession").(common.AuthSession)
	if !ok || authSession.Token == "" {
		return nil, errs.ERR_EXPIRED_SESSION
	}

	res, err := mw.AuthService.ValidateToken(
		r.Context(),
		&pb.ValidateTokenRequest{
			Token: authSession.Token,
			Aud:   aud.String(),
		},
	)
	if err != nil {
		return nil, errs.ERR_INVALID_TOKEN
	}

	return &auth.ValidToken{
		UserID:      res.UserId,
		TokenString: res.TokenString,
	}, nil
}
