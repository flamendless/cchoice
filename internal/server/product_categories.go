package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func AddProductCategoriesHandlers(s *Server, r chi.Router) {
	r.Get("/product-categories/side-panel/list", s.categoriesSidePanelHandler)
	r.Get("/product-categories/sections", s.categorySectionHandler)
	r.Get("/product-categories/{category_id}/products", s.categoryProductsHandler)
}

func (s *Server) categoriesSidePanelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Categories Side Panel Handler]"
	ctx := r.Context()

	categories, err := requests.GetCategoriesSidePanel(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		[]byte("key_categories_side_panel"),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := components.CategoriesSidePanelList(categories).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) categorySectionHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Categories Section Handler]"
	ctx := r.Context()

	page := 0
	if paramPage := r.URL.Query().Get("page"); paramPage != "" {
		if parsed, err := strconv.Atoi(paramPage); err == nil {
			page = parsed
		}
	}

	limit := constants.DefaultLimitCategories
	if paramLimit := r.URL.Query().Get("limit"); paramLimit != "" {
		if parsed, err := strconv.Atoi(paramLimit); err == nil {
			limit = max(parsed, constants.DefaultLimitCategories)
		}
	}

	res, err := requests.GetCategorySectionHandler(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		s.encoder,
		fmt.Appendf([]byte{}, "categorySectionHandler_p%d_l%d", page, limit),
		page,
		limit,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := components.CategorySection(page, res).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) categoryProductsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Category Products Handler]"
	ctx := r.Context()

	categoryID := chi.URLParam(r, "category_id")
	if categoryID == "" {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	categoryDBID := s.encoder.Decode(categoryID)
	category, err := s.dbRO.GetQueries().GetProductCategoryByID(ctx, categoryDBID)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if category.Category.String == "" {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.Int64("category id", category.ID),
			zap.String("subcategory", category.Subcategory.String),
		)
		return
	}

	products, err := s.dbRO.GetQueries().GetProductsByCategoryID(ctx, queries.GetProductsByCategoryIDParams{
		CategoryID: categoryDBID,
		Limit:      constants.DefaultLimitProducts,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(products) == 0 {
		logs.LogCtx(ctx).Debug(
			logtag,
			zap.Int64("category id", category.ID),
			zap.String("category name", category.Category.String),
		)
		return
	}

	validProducts := make([]int, 0, len(products))
	for i, product := range products {
		if !strings.HasSuffix(product.ThumbnailPath, constants.EmptyImageFilename) {
			validProducts = append(validProducts, i)
		} else {
			logs.LogCtx(ctx).Debug(
				"No valid image/thumbnail",
				zap.Int64("product id", product.ID),
			)
		}
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	var mu sync.Mutex
	productsWithValidImages := make([]queries.GetProductsByCategoryIDRow, 0, len(validProducts))

	for _, i := range validProducts {
		g.Go(func() error {
			imgData, err := images.GetImageDataB64(s.cache, s.productImageFS, products[i].ThumbnailPath, images.IMAGE_FORMAT_WEBP)
			if err != nil {
				logs.LogCtx(gctx).Error(
					logtag,
					zap.String("thumbnailPath", products[i].ThumbnailPath),
					zap.Error(err),
				)
				return nil
			}

			mu.Lock()
			products[i].ThumbnailData = imgData
			productsWithValidImages = append(productsWithValidImages, products[i])
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, "Failed to load images", http.StatusInternalServerError)
		return
	}

	categorySectionProducts := models.CategorySectionProducts{
		ID:          categoryID,
		Category:    utils.SlugToTile(category.Category.String),
		Subcategory: utils.SlugToTile(category.Subcategory.String),
		Products:    models.ToCategorySectionProducts(s.encoder, productsWithValidImages),
	}

	if err := components.CategorySectionProducts(categorySectionProducts).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
