package server

import (
	"fmt"
	"net/http"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/cart"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddCartsHandlers(s *Server, r chi.Router) {
	r.Get("/carts", s.cartsPageHandler)
	r.Get("/carts/summary-bar", s.getCartSummaryBarHandler)
	r.Get("/carts/lines", s.cartLinesHandler)
	r.Get("/carts/lines/count", s.getCartLinesCountHandler)
	r.Post("/carts/lines", s.addProductToCartHandler)
	r.Delete("/carts/lines/{checkoutline_id}", s.removeProductFromCartHandler)
	r.Patch("/carts/lines/{checkoutline_id}", s.updateCartLinesQtyHandler)
}

func (s *Server) cartsPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Page Handler]"
	checkoutlineProductIDs, ok := s.sessionManager.Get(r.Context(), skCheckoutLineProductIDs).([]string)
	if len(checkoutlineProductIDs) == 0 {
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := s.sessionManager.Token(r.Context())
	if !ok {
		logs.Log().Fatal(
			logtag,
			zap.Error(errs.ErrSessionCheckoutLineProductIDs),
			zap.String("token", token),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
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
		logs.Log().Fatal(logtag, zap.Error(err), zap.String("token", token))
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := components.CartPage(components.CartPageBody()).Render(r.Context(), w); err != nil {
		logs.Log().Fatal(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) cartLinesHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Lines Handler]"
	token := s.sessionManager.Token(r.Context())
	checkoutLines, err := cart.GetCheckoutLines(r.Context(), s.dbRO, token)
	if err != nil {
		logs.Log().Warn(logtag, zap.Error(err), zap.String("token", token))
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	for _, checkoutLine := range checkoutLines {
		var imgData string
		if checkoutLine.ThumbnailPath != constants.PathEmptyImage {
			finalPath, ext, err := images.GetImagePathWithSize(
				checkoutLine.ThumbnailPath,
				constants.CartPageThumbnailSize,
				true,
			)
			if err == nil {
				if imgDataB64, err := images.GetImageDataB64(s.cache, s.fs, finalPath, ext); err == nil {
					imgData = imgDataB64
				}
			}
		}

		price := utils.NewMoney(checkoutLine.UnitPriceWithVat, checkoutLine.UnitPriceWithVatCurrency)

		//TODO: (Brandon) - Discounts/sales
		discountedPrice := utils.NewMoney(checkoutLine.UnitPriceWithVat, checkoutLine.UnitPriceWithVatCurrency)

		cl := models.CheckoutLine{
			ID:              s.encoder.Encode(checkoutLine.ID),
			CheckoutID:      s.encoder.Encode(checkoutLine.CheckoutID),
			ProductID:       s.encoder.Encode(checkoutLine.ProductID),
			Name:            checkoutLine.Name,
			BrandName:       checkoutLine.BrandName,
			Quantity:        checkoutLine.Quantity,
			ThumbnailPath:   checkoutLine.ThumbnailPath,
			ThumbnailData:   imgData,
			Price:           *price,
			DiscountedPrice: *discountedPrice,
			Total:           *discountedPrice.Multiply(checkoutLine.Quantity),
		}

		if err := components.CartCheckoutLineItem(cl).Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}
	}
}

func (s *Server) addProductToCartHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Add Product To Cart Handler]"
	if err := r.ParseForm(); err != nil {
		logs.Log().Fatal(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	productID := r.Form.Get("product_id")
	if productID == "" {
		logs.Log().Fatal(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	dbProductID := s.encoder.Decode(productID)
	if _, err := s.dbRO.GetQueries().CheckProductExistsByID(r.Context(), dbProductID); err != nil {
		logs.Log().Fatal(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkoutLineProductIDs, err := AddToCheckoutLineProductIDs(r.Context(), s.sessionManager, productID)
	if err != nil {
		logs.Log().Fatal(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("product id", productID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logs.Log().Info(
		logtag,
		zap.String("token", s.sessionManager.Token(r.Context())),
		zap.Strings("checkout line product ids", checkoutLineProductIDs),
	)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) removeProductFromCartHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Remove Product From Cart Handler]"
	checkoutLineID := chi.URLParam(r, "checkoutline_id")
	dbCheckoutLineID := s.encoder.Decode(checkoutLineID)

	checkoutLine, err := s.dbRO.GetQueries().GetCheckoutLineByID(r.Context(), dbCheckoutLineID)
	if err != nil {
		logs.Log().Fatal(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := RemoveFromCheckoutLineProductIDs(
		r.Context(),
		s.sessionManager,
		s.encoder.Encode(checkoutLine.ProductID),
	); err != nil {
		logs.Log().Fatal(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.dbRW.GetQueries().DeleteCheckoutLineByID(r.Context(), dbCheckoutLineID); err != nil {
		logs.Log().Fatal(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	remaining, err := s.dbRO.GetQueries().CountCheckoutLineByCheckoutID(r.Context(), checkoutLine.CheckoutID)
	if err != nil {
		logs.Log().Fatal(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if remaining == 0 {
		w.Header().Set("HX-Redirect", "/cchoice/carts")
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getCartSummaryBarHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Get Cart Summary Bar Handler]"
	data := r.URL.Query().Get("data")
	if data == "" {
		logs.Log().Fatal(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	if data == "items" {
		token := s.sessionManager.Token(r.Context())
		checkoutLines, err := cart.GetCheckoutLines(r.Context(), s.dbRO, token)
		if err != nil {
			if _, err := w.Write([]byte("0 Items")); err != nil {
				logs.Log().Fatal(logtag, zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		count := 0
		for _, checkoutLine := range checkoutLines {
			count += int(checkoutLine.Quantity)
		}
		if _, err := w.Write(fmt.Appendf(nil, "%d Items", count)); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) getCartLinesCountHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Get Cart Lines Count Handler]"
	count := 0
	if productIDs, ok := s.sessionManager.Get(r.Context(), skCheckoutLineProductIDs).([]string); ok {
		count = len(productIDs)
	}
	if _, err := w.Write(fmt.Appendf(nil, "%d", count)); err != nil {
		logs.Log().Fatal(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) updateCartLinesQtyHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Update Cart Lines Qty Handler]"
	checkoutLineID := chi.URLParam(r, "checkoutline_id")
	dbCheckoutLineID := s.encoder.Decode(checkoutLineID)

	qty := 0
	if r.URL.Query().Get("dec") == "1" {
		qty = -1
	} else if r.URL.Query().Get("inc") == "1" {
		qty = 1
	}

	if qty == 0 {
		logs.Log().Fatal(
			logtag,
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	newQty, _ := s.dbRW.GetQueries().UpdateCheckoutLineQtyByID(
		r.Context(),
		queries.UpdateCheckoutLineQtyByIDParams{
			ID:       dbCheckoutLineID,
			Quantity: int64(qty),
		},
	)
	if newQty == 0 {
		if r.URL.Query().Get("dec") == "1" {
			newQty = 1
		} else if r.URL.Query().Get("inc") == "1" {
			newQty = constants.MaxCartLineQty
		}
	}

	if _, err := w.Write(fmt.Appendf(nil, "Qty: %d", newQty)); err != nil {
		logs.Log().Fatal(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
