package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	pb "cchoice/proto"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type ShopService interface {
	pb.ShopServiceClient
}

type ShopHandler struct {
	Logger      *zap.Logger
	ShopService ShopService
	SM          *scs.SessionManager
}

func NewShopHandler(
	logger *zap.Logger,
	shopService ShopService,
	sm *scs.SessionManager,
) ShopHandler {
	return ShopHandler{
		Logger:      logger,
		ShopService: shopService,
		SM:          sm,
	}
}

func (h ShopHandler) HomePage(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	return &common.HandlerRes{
		Component: components.ShopHome(),
		RedirectTo: "/home",
	}
}
