package server

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
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

	productsService := services.NewProductsService(s.dbRO)
	isUnique, err := productsService.ValidateSerial(ctx, serial)
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

func (s *Server) adminSuperuserProductsPostHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Post Handler]"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to parse form"))
		return
	}

	brandIDStr := r.FormValue("brand_id")
	if brandIDStr == "" {
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Brand is required"))
		return
	}
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Invalid brand"))
		return
	}

	_, err = s.dbRO.GetQueries().GetBrandsByID(ctx, brandID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Brand not found"))
		return
	}

	category := r.FormValue("category")
	subcategory := r.FormValue("subcategory")
	if category == "" || subcategory == "" {
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Category and subcategory are required"))
		return
	}

	categoryRow, err := s.dbRO.GetQueries().GetProductCategoryByCategoryAndSubcategory(ctx, queries.GetProductCategoryByCategoryAndSubcategoryParams{
		Category:    sql.NullString{String: category, Valid: true},
		Subcategory: sql.NullString{String: subcategory, Valid: true},
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Category not found"))
		return
	}

	name := r.FormValue("name")
	serial := r.FormValue("serial")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")

	if name == "" || serial == "" || description == "" || priceStr == "" {
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "All fields are required"))
		return
	}

	productsService := services.NewProductsService(s.dbRO)
	isUnique, err := productsService.ValidateSerial(ctx, serial)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to validate serial"))
		return
	}

	if !isUnique {
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Serial already exists"))
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Invalid price"))
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
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "All product specs are required"))
		return
	}

	file, header, err := r.FormFile("product_image")
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Product image is required"))
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	uuid := utils.GenString(16)
	filename := fmt.Sprintf("products/%s_%s%s", uuid, strings.ReplaceAll(name, " ", "_"), ext)

	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, file); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to read image"))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if err := s.objectStorage.PutObject(ctx, filename, &buf, contentType); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to upload image"))
		return
	}

	now := time.Now().UTC()

	specs, err := s.dbRW.GetQueries().CreateProductSpecs(ctx, queries.CreateProductSpecsParams{
		Colours:       sql.NullString{String: specColours, Valid: true},
		Sizes:         sql.NullString{String: specSizes, Valid: true},
		Segmentation:  sql.NullString{String: specSegmentation, Valid: true},
		PartNumber:    sql.NullString{String: specPartNumber, Valid: true},
		Power:         sql.NullString{String: specPower, Valid: true},
		Capacity:      sql.NullString{String: specCapacity, Valid: true},
		ScopeOfSupply: sql.NullString{String: specScopeOfSupply, Valid: true},
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to create product specs"))
		return
	}

	product, err := s.dbRW.GetQueries().CreateProducts(ctx, queries.CreateProductsParams{
		Serial:                      serial,
		Name:                        name,
		Description:                 sql.NullString{String: description, Valid: true},
		BrandID:                     brandID,
		Status:                      enums.PRODUCT_STATUS_ACTIVE.String(),
		ProductSpecsID:              sql.NullInt64{Int64: specs.ID, Valid: true},
		UnitPriceWithoutVat:         unitPriceWithoutVat,
		UnitPriceWithVat:            unitPriceWithVat,
		UnitPriceWithoutVatCurrency: "PHP",
		UnitPriceWithVatCurrency:    "PHP",
		CreatedAt:                   now,
		UpdatedAt:                   now,
		DeletedAt:                   constants.DtBeginning,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to create product"))
		return
	}

	_, err = s.dbRW.GetQueries().CreateProductsCategories(ctx, queries.CreateProductsCategoriesParams{
		ProductID:  product.ID,
		CategoryID: categoryRow.ID,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to link product to category"))
		return
	}

	_, err = s.dbRW.GetQueries().CreateProductImage(ctx, queries.CreateProductImageParams{
		ProductID: product.ID,
		Path:      filename,
		Thumbnail: sql.NullString{String: filename, Valid: true},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: constants.DtBeginning,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", "Failed to save product image"))
		return
	}

	logs.LogCtx(ctx).Info(logtag, zap.Int64("product_id", product.ID), zap.String("name", name))
	redirectHX(w, r, utils.URLWithSuccess("/admin/superuser/products", "Product created successfully"))
}

func (s *Server) adminSuperuserProductsHandler(w http.ResponseWriter, r *http.Request) {
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
			ID:   b.ID,
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
