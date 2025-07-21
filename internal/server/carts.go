package server

import (
	"fmt"
	"net/http"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/cart"
	"cchoice/internal/errs"
	"cchoice/internal/logs"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddCartsHandlers(s *Server, r chi.Router) {
	r.Get("/carts", s.cartsPageHandler)
	r.Get("/carts/lines", s.cartLinesHandler)
	r.Get("/carts/lines/count", s.getCartLinesCountHandler)
	r.Post("/carts/lines", s.addProductToCartHandler)
}

func (s *Server) cartsPageHandler(w http.ResponseWriter, r *http.Request) {
	checkoutlineProductIDs, ok := s.sessionManager.Get(r.Context(), skCheckoutLineProductIDs).([]string)
	if len(checkoutlineProductIDs) == 0 {
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal("Cart page handler", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := s.sessionManager.Token(r.Context())
	if !ok {
		logs.Log().Fatal(
			"Cart page handler",
			zap.Error(errs.ERR_SESSION_CHECKOUT_LINE_PRODUCT_IDS),
			zap.String("token", token),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal("Cart page handler", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	checkoutID, err := cart.CreateCart(
		r.Context(),
		s.dbRW.GetQueries(),
		s.encoder,
		token,
		checkoutlineProductIDs,
	)
	if err != nil || checkoutID == -1 {
		logs.Log().Fatal(
			"Cart page handler",
			zap.Error(err),
			zap.String("token", token),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal("Cart page handler", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := components.CartPage(components.CartPageBody()).Render(r.Context(), w); err != nil {
		logs.Log().Fatal("Cart page handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) cartLinesHandler(w http.ResponseWriter, r *http.Request) {
	token := s.sessionManager.Token(r.Context())
	checkoutID, err := s.dbRO.GetQueries().GetCheckoutIDBySessionID(r.Context(), token)
	if err != nil {
		logs.Log().Warn(
			"Carts lines handler",
			zap.Error(err),
			zap.String("token", token),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal("Cart page handler", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	checkoutLines, err := s.dbRO.GetQueries().GetCheckoutLinesByCheckoutID(r.Context(), checkoutID)
	if err != nil || len(checkoutLines) == 0 {
		logs.Log().Warn(
			"Carts lines handler",
			zap.Error(err),
			zap.String("token", token),
			zap.Int("checkout lines", len(checkoutLines)),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal("Cart page handler", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	for _, checkoutLine := range checkoutLines {
		cl := models.CheckoutLine{
			ID:         s.encoder.Encode(checkoutLine.ID),
			CheckoutID: s.encoder.Encode(checkoutLine.CheckoutID),
			ProductID:  s.encoder.Encode(checkoutLine.ProductID),
			Name:       checkoutLine.Name,
			BrandName:  checkoutLine.BrandName,
			Quantity:   checkoutLine.Quantity,
		}

		if err := components.CartCheckoutLineItem(cl).Render(r.Context(), w); err != nil {
			logs.Log().Fatal("Cart lines handler", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}
	}
}

func (s *Server) addProductToCartHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		logs.Log().Fatal("checkouts lines handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	productID := r.Form.Get("product_id")
	if productID == "" {
		err := errs.ERR_INVALID_PARAMS
		logs.Log().Fatal("checkouts lines handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbProductID := s.encoder.Decode(productID)
	if _, err := s.dbRO.GetQueries().CheckProductExistsByID(r.Context(), dbProductID); err != nil {
		logs.Log().Fatal(
			"checkouts lines handler",
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkoutLineProductIDs, err := AddToCheckoutLineProductIDs(r.Context(), s.sessionManager, productID)
	if err != nil {
		logs.Log().Fatal(
			"add checkout line",
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("product id", productID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logs.Log().Info(
		"checkouts lines",
		zap.String("token", s.sessionManager.Token(r.Context())),
		zap.Strings("checkout line product ids", checkoutLineProductIDs),
	)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getCartLinesCountHandler(w http.ResponseWriter, r *http.Request) {
	count := 0
	if productIDs, ok := s.sessionManager.Get(r.Context(), skCheckoutLineProductIDs).([]string); ok {
		count = len(productIDs)
	}
	w.Write(fmt.Appendf(nil, "%d", count))
}
