package server

import (
	"net/http"
	"strconv"
	"strings"

	compshop "cchoice/cmd/web/components/shop"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddSearchHandlers(s *Server, r chi.Router) {
	r.Get("/search", s.searchPageHandler)
	r.Get("/search/products", s.searchProductsHandler)
	r.Get("/search/related", s.searchRelatedProductsHandler)
}

func (s *Server) searchPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Search Page Handler]"
	ctx := r.Context()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < constants.MinSearchQueryLength {
		logs.LogCtx(ctx).Warn(logtag, zap.String("query", query))
		http.Redirect(w, r, utils.URL("/"), http.StatusSeeOther)
		return
	}

	if err := compshop.SearchPage(models.SearchPageData{Query: query}).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) searchProductsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Search Products Handler]"
	ctx := r.Context()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < constants.MinSearchQueryLength {
		logs.LogCtx(ctx).Warn(logtag, zap.String("query", query))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	page := 0
	if paramPage := r.URL.Query().Get("page"); paramPage != "" {
		if parsed, err := strconv.Atoi(paramPage); err == nil && parsed >= 0 {
			page = parsed
		}
	}

	limit := constants.DefaultLimitSearchResultsPage
	offset := page * limit

	rows, err := s.dbRO.GetQueries().GetProductsBySearchQueryPaginated(ctx, queries.GetProductsBySearchQueryPaginatedParams{
		Name:   query,
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("query", query))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	validRows := filterSearchPaginatedRows(rows)
	if len(validRows) == 0 {
		if page == 0 {
			if err := compshop.SearchNoResults(query).Render(ctx, w); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if err := compshop.SearchProductsExhausted(query).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	hasMore := len(validRows) == limit
	products := models.ToProductGridProductsFromSearchRows(s.encoder, s.GetCDNURL, validRows)

	if err := compshop.SearchProductsPage(models.SearchProductsPageData{
		Query:    query,
		Page:     page,
		HasMore:  hasMore,
		Products: products,
	}).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const searchRelatedSource = "related"
const searchOtherSource = "other"

func (s *Server) searchRelatedProductsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Search Related Products Handler]"
	ctx := r.Context()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < constants.MinSearchQueryLength {
		logs.LogCtx(ctx).Warn(logtag, zap.String("query", query))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	page := 0
	if paramPage := r.URL.Query().Get("page"); paramPage != "" {
		if parsed, err := strconv.Atoi(paramPage); err == nil && parsed >= 0 {
			page = parsed
		}
	}

	source := r.URL.Query().Get("source")
	if source != searchRelatedSource && source != searchOtherSource {
		source = searchRelatedSource
	}

	limit := constants.DefaultLimitSearchResultsPage
	offset := page * limit

	var products []models.CategorySectionProduct
	hasMore := false

	switch source {
	case searchOtherSource:
		otherRows, err := s.dbRO.GetQueries().GetOtherProductsForSearch(ctx, queries.GetOtherProductsForSearchParams{
			Name:   query,
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("query", query))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		validOtherRows := filterOtherSearchRows(otherRows)
		hasMore = len(validOtherRows) == limit
		products = models.ToProductGridProductsFromOtherRows(s.encoder, s.GetCDNURL, validOtherRows)
	default:
		relatedRows, err := s.dbRO.GetQueries().GetRelatedProductsForSearch(ctx, queries.GetRelatedProductsForSearchParams{
			Name:   query,
			Name_2: query,
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("query", query))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		validRows := filterRelatedSearchRows(relatedRows)
		if page == 0 && len(validRows) == 0 {
			source = searchOtherSource

			otherRows, otherErr := s.dbRO.GetQueries().GetOtherProductsForSearch(ctx, queries.GetOtherProductsForSearchParams{
				Name:   query,
				Limit:  int64(limit),
				Offset: int64(offset),
			})
			if otherErr != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(otherErr), zap.String("query", query))
				http.Error(w, otherErr.Error(), http.StatusInternalServerError)
				return
			}

			validOtherRows := filterOtherSearchRows(otherRows)
			hasMore = len(validOtherRows) == limit
			products = models.ToProductGridProductsFromOtherRows(s.encoder, s.GetCDNURL, validOtherRows)
			break
		}

		hasMore = len(validRows) == limit
		products = models.ToProductGridProductsFromRelatedRows(s.encoder, s.GetCDNURL, validRows)
	}

	if len(products) == 0 {
		return
	}

	if err := compshop.SearchRelatedProductsPage(models.SearchRelatedProductsPageData{
		Query:    query,
		Page:     page,
		HasMore:  hasMore,
		Source:   source,
		Products: products,
	}).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func filterSearchPaginatedRows(rows []queries.GetProductsBySearchQueryPaginatedRow) []queries.GetProductsBySearchQueryPaginatedRow {
	valid := make([]queries.GetProductsBySearchQueryPaginatedRow, 0, len(rows))
	for _, row := range rows {
		if strings.HasSuffix(row.ThumbnailPath, constants.EmptyImageFilename) {
			continue
		}
		if !row.Slug.Valid || row.Slug.String == "" {
			continue
		}
		valid = append(valid, row)
	}
	return valid
}

func filterRelatedSearchRows(rows []queries.GetRelatedProductsForSearchRow) []queries.GetRelatedProductsForSearchRow {
	valid := make([]queries.GetRelatedProductsForSearchRow, 0, len(rows))
	for _, row := range rows {
		if strings.HasSuffix(row.ThumbnailPath, constants.EmptyImageFilename) {
			continue
		}
		if !row.Slug.Valid || row.Slug.String == "" {
			continue
		}
		valid = append(valid, row)
	}
	return valid
}

func filterOtherSearchRows(rows []queries.GetOtherProductsForSearchRow) []queries.GetOtherProductsForSearchRow {
	valid := make([]queries.GetOtherProductsForSearchRow, 0, len(rows))
	for _, row := range rows {
		if strings.HasSuffix(row.ThumbnailPath, constants.EmptyImageFilename) {
			continue
		}
		if !row.Slug.Valid || row.Slug.String == "" {
			continue
		}
		valid = append(valid, row)
	}
	return valid
}
