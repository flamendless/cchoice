package server

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminSuperuserProductsSubcategoriesHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Subcategories Handler]"
	ctx := r.Context()

	category := r.URL.Query().Get("category")
	if category == "" {
		logs.LogCtx(ctx).Error(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	categories, err := requests.GetCategoriesForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminCategoriesCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	subcategories, exists := categories[category]
	if !exists {
		logs.LogCtx(ctx).Error(logtag, zap.String("category", category))
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}

	if err := compadmin.SubcategoryOptions(subcategories).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserProductsValidateSerialHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Validate Serial Handler]"
	ctx := r.Context()

	serial := r.URL.Query().Get("serial")
	if serial == "" {
		if err := compadmin.SerialValidationResult(false).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		}
		return
	}

	isUnique, err := s.services.products.ValidateSerial(ctx, serial)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		if err := compadmin.SerialValidationResult(false).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		}
		return
	}

	if !isUnique {
		if err := compadmin.SerialValidationResult(false).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		}
		return
	}

	if err := compadmin.SerialValidationResult(true).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func (s *Server) adminSuperuserProductsCreatePostHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Post Handler]"
	const page = "/admin/superuser/products/create"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	brandID := s.encoder.Decode(r.FormValue("brand_id"))
	if brandID == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(page, "Invalid brand"))
		return
	}

	category := r.FormValue("category")
	subcategory := r.FormValue("subcategory")
	if category == "" || subcategory == "" {
		redirectHX(w, r, utils.URLWithError(page, "Category and subcategory are required"))
		return
	}

	name := r.FormValue("name")
	serial := r.FormValue("serial")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")
	if name == "" || serial == "" || description == "" || priceStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	isUnique, err := s.services.products.ValidateSerial(ctx, serial)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to validate serial"))
		return
	}

	if !isUnique {
		redirectHX(w, r, utils.URLWithError(page, "Serial already exists"))
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid price"))
		return
	}

	vatPercentage, err := strconv.ParseFloat(conf.Conf().Settings.VATPercentage, 64)
	if err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
		vatPercentage = 0
	}
	unitPriceWithoutVat := int64(math.Round(price / (1 + vatPercentage/100)))
	unitPriceWithVat := int64(math.Round(price))

	specColours := r.FormValue("spec_colours")
	specSizes := r.FormValue("spec_sizes")
	specSegmentation := r.FormValue("spec_segmentation")
	specPartNumber := r.FormValue("spec_part_number")
	specPower := r.FormValue("spec_power")
	specCapacity := r.FormValue("spec_capacity")
	specScopeOfSupply := r.FormValue("spec_scope_of_supply")

	if specColours == "" || specSizes == "" || specSegmentation == "" ||
		specPartNumber == "" || specPower == "" || specCapacity == "" || specScopeOfSupply == "" {
		redirectHX(w, r, utils.URLWithError(page, "All product specs are required"))
		return
	}

	file, header, err := r.FormFile("product_image")
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Product image is required"))
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	uuid := utils.GenString(16)
	filename := fmt.Sprintf("products/%s_%s%s", uuid, strings.ReplaceAll(name, " ", "_"), ext)

	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, file); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to read image"))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if err := s.objectStorage.PutObject(ctx, filename, &buf, contentType); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to upload image"))
		return
	}

	product, err := s.services.products.CreateProduct(ctx, services.CreateProductInput{
		Serial:      serial,
		Name:        name,
		Description: description,
		BrandID:     brandID,
		Category:    category,
		Subcategory: subcategory,
		Specs: services.ProductSpecsInput{
			Colours:       specColours,
			Sizes:         specSizes,
			Segmentation:  specSegmentation,
			PartNumber:    specPartNumber,
			Power:         specPower,
			Capacity:      specCapacity,
			ScopeOfSupply: specScopeOfSupply,
		},
		ImagePath:           filename,
		UnitPriceWithoutVat: unitPriceWithoutVat,
		UnitPriceWithVat:    unitPriceWithVat,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to create product"))
		return
	}

	logs.LogCtx(ctx).Info(logtag, zap.Int64("product_id", product.ID), zap.String("name", name))
	redirectHX(w, r, utils.URLWithSuccess(page, "Product created successfully"))
}

func (s *Server) adminSuperuserProductsCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Handler]"
	ctx := r.Context()

	brandsRes, err := requests.GetBrandsForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminBrandsCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		brandsRes = []queries.GetBrandsForSidePanelRow{}
	}

	brands := make([]models.AdminBrand, 0, len(brandsRes))
	for _, b := range brandsRes {
		brands = append(brands, models.AdminBrand{
			ID:   s.encoder.Encode(b.ID),
			Name: b.Name,
		})
	}

	categoriesRes, err := requests.GetCategoriesForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminCategoriesCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		categoriesRes = map[string][]string{}
	}

	categories := make([]models.AdminCategory, 0, len(categoriesRes))
	for cat, subcats := range categoriesRes {
		categories = append(categories, models.AdminCategory{
			Category:      cat,
			Subcategories: subcats,
		})
	}

	formData := models.AdminProductForm{
		Brands:        brands,
		Categories:    categories,
		VATPercentage: conf.Conf().Settings.VATPercentage,
	}

	if err := compadmin.AdminSuperuserProductsPage(formData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
