package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cchoice/internal/logs"
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

	sitemapLine := fmt.Sprintf("Sitemap: %s", utils.SiteURL("/sitemap.xml"))
	body := strings.Replace(string(content), "# __SITEMAP__", sitemapLine, 1)

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

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

	writeSitemapURL(&buf, utils.SiteURL("/"), time.Now().UTC(), "daily", "1.0")

	for _, product := range products {
		if !product.Slug.Valid || product.Slug.String == "" {
			continue
		}
		lastMod := product.UpdatedAt
		if lastMod.IsZero() {
			lastMod = time.Now().UTC()
		}
		writeSitemapURL(
			&buf,
			utils.SiteURL("/product/"+product.Slug.String),
			lastMod.UTC(),
			"weekly",
			"0.8",
		)
	}

	buf.WriteString(`</urlset>`)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if _, err := w.Write(buf.Bytes()); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func writeSitemapURL(buf *bytes.Buffer, loc string, lastMod time.Time, changefreq string, priority string) {
	fmt.Fprintf(buf, "<url><loc>%s</loc>", escapeXML(loc))
	fmt.Fprintf(buf, "<lastmod>%s</lastmod>", lastMod.Format("2006-01-02"))
	fmt.Fprintf(buf, "<changefreq>%s</changefreq>", changefreq)
	fmt.Fprintf(buf, "<priority>%s</priority></url>", priority)
}

func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"'", "&apos;",
		"\"", "&quot;",
	)
	return replacer.Replace(s)
}
