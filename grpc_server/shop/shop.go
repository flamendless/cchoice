package shop

import (
	"cchoice/internal/ctx"
	pb "cchoice/proto"
)

type ShopServer struct {
	pb.UnimplementedShopServiceServer
	CtxDB *ctx.Database
}

func NewGRPCShopServer(ctxDB *ctx.Database) *ShopServer {
	return &ShopServer{CtxDB: ctxDB}
}
