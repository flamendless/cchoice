package settings

import (
	"cchoice/internal/ctx"
	"cchoice/internal/errs"
	pb "cchoice/proto"
	"context"
)

type SettingsServer struct {
	pb.UnimplementedSettingsServiceServer
	CtxDB *ctx.Database
}

func NewGRPCSettingsServer(ctxDB *ctx.Database) *SettingsServer {
	return &SettingsServer{CtxDB: ctxDB}
}

func (s *SettingsServer) GetSettingsByNames(
	ctx context.Context,
	in *pb.SettingsByNamesRequest,
) (*pb.SettingsResponse, error) {
	res, err := s.CtxDB.QueriesRead.GetSettingsByNames(context.TODO(), in.Names)
	if err != nil {
		return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
	}

	settings := map[string]string{}
	for _, setting := range res {
		settings[setting.Name] = setting.Value
	}

	return &pb.SettingsResponse{
		Length:   int64(len(settings)),
		Settings: settings,
	}, nil
}
