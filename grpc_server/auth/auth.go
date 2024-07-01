package auth

import (
	"cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	CtxDB  *ctx.Database
	Issuer *auth.Issuer
}

func (s *AuthServer) Authenticate(
	ctx context.Context,
	in *pb.AuthLoginRequest,
) (*pb.AuthLoginResponse, error) {
	username := in.GetUsername()
	password := in.GetPassword()

	errUsername := utils.ValidateUsername(username)
	if errUsername != nil {
		return nil, errUsername
	}

	errPW := utils.ValidatePW(password)
	if errPW != nil {
		return nil, errPW
	}

	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password: %w", err)
	}

	hashedDBPassword, err := s.CtxDB.QueriesRead.GetUserHashedPassword(context.Background(), username)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	match, err := argon2id.ComparePasswordAndHash(hashedDBPassword, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	if !match {
		return nil, fmt.Errorf("Invalid credentials")
	}

	token, err := s.Issuer.IssueToken(enums.AudAPI, username)
	if err != nil {
		return nil, errPW
	}

	res := &pb.AuthLoginResponse{
		Token: token,
	}

	return res, nil
}
