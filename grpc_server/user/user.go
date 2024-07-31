package user

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/serialize"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	CtxDB *ctx.Database
}

func NewGRPCUserServer(ctxDB *ctx.Database) *UserServer {
	return &UserServer{CtxDB: ctxDB}
}

func (s *UserServer) Register(
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

	userID, err := s.CtxDB.Queries.CreateUser(context.Background(), cchoice_db.CreateUserParams{
		FirstName:  in.FirstName,
		MiddleName: in.MiddleName,
		LastName:   in.LastName,
		Email:      in.Email,
		Password:   hashedPassword,
		MobileNo:   in.MobileNo,
		UserType:   enums.USER_TYPE_API.String(),
		Status:     pb.UserStatus_ACTIVE.String(),
	})
	if err != nil {
		return nil, err
	}

	err = s.CtxDB.Queries.CreateInitialAuth(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		UserId: serialize.EncDBID(userID),
		Token:  "",
	}, nil
}

func (s *UserServer) GetUserByID(
	ctx context.Context,
	in *pb.GetUserByIDRequest,
) (*pb.GetUserByIDResponse, error) {
	userID := serialize.DecDBID(in.UserId)
	res, err := s.CtxDB.QueriesRead.GetUserWithAuthByID(
		context.Background(),
		userID,
	)
	if err != nil {
		return nil, errs.ERR_INVALID_RESOURCE
	}

	return &pb.GetUserByIDResponse{
		User: &pb.User{
			Id:         in.UserId,
			FirstName:  res.FirstName,
			MiddleName: res.MiddleName,
			LastName:   res.LastName,
			Email:      res.Email,
			OtpEnabled: res.OtpEnabled,
		},
	}, nil
}
