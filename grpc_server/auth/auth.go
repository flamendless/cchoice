package auth

import (
	"cchoice/cchoice_db"
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
	in *pb.AuthenticateRequest,
) (*pb.AuthenticateResponse, error) {
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

	hashedDBPassword, err := s.CtxDB.QueriesRead.GetUserHashedPassword(context.Background(), username)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	match, err := argon2id.ComparePasswordAndHash(password, hashedDBPassword)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	if !match {
		return nil, fmt.Errorf("Invalid credentials")
	}

	tokenString, err := s.Issuer.IssueToken(enums.AUD_API, username)
	if err != nil {
		return nil, err
	}

	res := &pb.AuthenticateResponse{
		Token: tokenString,
	}

	return res, nil
}

func (s *AuthServer) Register(
	ctx context.Context,
	in *pb.RegisterRequest,
) (*pb.RegisterResponse, error) {
	err := utils.ValidateUserReg(in)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := argon2id.CreateHash(in.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password: %w", err)
	}

	err = s.CtxDB.Queries.CreateUser(context.Background(), cchoice_db.CreateUserParams{
		FirstName:  in.FirstName,
		MiddleName: in.MiddleName,
		LastName:   in.LastName,
		Email:      in.Email,
		Password:   hashedPassword,
		MobileNo:   in.MobileNo,
		UserType:   enums.USER_TYPE_API.String(),
		Status:     enums.PRODUCT_STATUS_ACTIVE.String(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		Token: "",
	}, nil
}
