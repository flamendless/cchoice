package server

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"cchoice/internal/logs"
	"cchoice/internal/seo"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddSEOHandlers(s *Server, r chi.Router) {
	r.Get("/robots.txt", s.robotsTxtHandler)
	r.Get("/sitemap.xml", s.sitemapHandler)
}

func (s *Server) robotsTxtHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Robots Handler]"
	ctx := r.Context()

	file, err := s.staticFS.Open("robots.txt")
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body := seo.InjectSitemapLine(string(content), utils.SiteURL("/sitemap.xml"))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if _, err := w.Write([]byte(body)); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func (s *Server) sitemapHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Sitemap Handler]"
	ctx := r.Context()

	products, err := s.services.product.ListActiveProductSlugs(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	entries := make([]seo.SitemapEntry, 0, len(products))
	for _, product := range products {
		if !product.Slug.Valid || product.Slug.String == "" {
			continue
		}
		lastMod := product.UpdatedAt
		if lastMod.IsZero() {
			lastMod = time.Now().UTC()
		}
		entries = append(entries, seo.SitemapEntry{
			Loc:        utils.SiteURL("/product/" + product.Slug.String),
			LastMod:    lastMod.UTC(),
			ChangeFreq: "weekly",
			Priority:   "0.8",
		})
	}

	categorySlugs, err := s.services.productCategory.ListCategorySitemapSlugs(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	parentCategories := make(map[string]struct{}, len(categorySlugs))
	for _, row := range categorySlugs {
		parentCategories[row.Category] = struct{}{}
		entries = append(entries, seo.SitemapEntry{
			Loc:        utils.SiteURL(fmt.Sprintf("/categories/%s/%s", row.Category, row.Subcategory)),
			LastMod:    now,
			ChangeFreq: "weekly",
			Priority:   "0.7",
		})
	}
	for category := range parentCategories {
		entries = append(entries, seo.SitemapEntry{
			Loc:        utils.SiteURL("/categories/" + category),
			LastMod:    now,
			ChangeFreq: "weekly",
			Priority:   "0.7",
		})
	}

	body := seo.BuildSitemapXML(utils.SiteURL("/"), entries)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if _, err := w.Write([]byte(body)); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}
