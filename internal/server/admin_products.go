package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
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

	isUnique, err := s.services.product.ValidateSerial(ctx, serial)
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

	brandID := r.FormValue("brand_id")
	if brandID == "" {
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

	isUnique, err := s.services.product.ValidateSerial(ctx, serial)
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
	specWeight := r.FormValue("spec_weight")
	specWeightUnit := r.FormValue("spec_weight_unit")

	if specColours == "" || specSizes == "" || specSegmentation == "" ||
		specPartNumber == "" || specPower == "" || specCapacity == "" || specScopeOfSupply == "" ||
		specWeight == "" || specWeightUnit == "" {
		redirectHX(w, r, utils.URLWithError(page, "All product specs are required"))
		return
	}

	result := "success"
	defer func() {
		if err := s.services.staffLog.CreateLog(
			context.Background(),
			s.sessionManager.GetString(ctx, SessionStaffID),
			constants.ActionCreate,
			constants.ModuleProducts,
			result,
			nil,
		); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
	}()

	var filename string
	var brandName string
	if conf.Conf().Test.LocalUploadImage || conf.Conf().IsProd() {
		file, header, err := r.FormFile("product_image")
		if err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, "Product image is required"))
			return
		}
		defer file.Close()

		brandName, err = s.services.brand.GetNameByID(ctx, brandID)
		if err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, "Brand not found"))
			return
		}

		filename = s.services.image.GenerateFilename(filepath.Ext(header.Filename), brandName, name)
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, file); err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, "Failed to read image"))
			return
		}

		contentType := header.Header.Get("Content-Type")
		if err := s.services.image.UploadProductImage(
			ctx,
			brandName,
			filename,
			&buf,
			contentType,
		); err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, "Failed to upload image"))
			return
		}
	}

	product, err := s.services.product.Create(ctx, services.CreateProductInput{
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
			Weight:        specWeight,
			WeightUnit:    enums.ParseWeightUnitToEnum(specWeightUnit).ToDB(),
		},
		ImagePath:           filename,
		UnitPriceWithoutVat: unitPriceWithoutVat,
		UnitPriceWithVat:    unitPriceWithVat,
	})
	if err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to create product"))
		return
	}

	if filename != "" && s.thumbnailJobRunner != nil {
		if err := s.thumbnailJobRunner.QueueThumbnailJob(ctx, jobs.ThumbnailJobParams{
			ProductID:  product.ID,
			Brand:      brandName,
			SourcePath: filename,
			Filename:   filepath.Base(filename),
		}); err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		}
	}

	result = fmt.Sprintf("success. ID '%s'", s.encoder.Encode(product.ID))

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

func (s *Server) adminSuperuserProductsListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminSuperuserProductsListPage("Products").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserProductsListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products List Table Handler]"
	ctx := r.Context()

	search := r.URL.Query().Get("search")
	statusStr := r.URL.Query().Get("status")
	status := enums.ParseProductStatusToEnum(statusStr)
	if statusStr != "" && status == enums.PRODUCT_STATUS_UNDEFINED {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("search", search),
			zap.String("status", statusStr),
			zap.Error(errs.ErrEnumInvalid),
		)
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", errs.ErrEnumInvalid.Error()))
		return
	}

	productList, err := s.services.product.GetForListingAdmin(ctx, search, status)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		productList = []models.AdminProductListItem{}
	}

	if err := compadmin.AdminSuperuserProductsListTable(productList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("search", search),
			zap.String("status", statusStr),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError("/admin/superuser/products", err.Error()))
		return
	}
}

func (s *Server) adminSuperuserProductsUpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Update Status Handler]"
	const page = "/admin/superuser/products"
	ctx := r.Context()

	productIDStr := chi.URLParam(r, "id")
	if productIDStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "Invalid product ID"))
		return
	}

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	statusStr := r.FormValue("status")
	if statusStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "Status is required"))
		return
	}

	status := enums.ParseProductStatusToEnum(statusStr)
	if status == enums.PRODUCT_STATUS_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, "Invalid status"))
		return
	}

	result := "success"
	defer func() {
		if err := s.services.staffLog.CreateLog(
			context.Background(),
			s.sessionManager.GetString(ctx, SessionStaffID),
			constants.ActionUpdateStatus,
			constants.ModuleProducts,
			result,
			nil,
		); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
	}()

	if err := s.services.product.UpdateStatus(ctx, productIDStr, status); err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("product_id", productIDStr),
			zap.String("status", statusStr),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError(page, "Failed to update product status"))
		return
	}

	result = fmt.Sprintf("success. ID '%s'", productIDStr)
	redirectHX(w, r, utils.URLWithSuccess(page, "Product status updated successfully"))
}

func (s *Server) adminSuperuserProductsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Delete Handler]"
	const page = "/admin/superuser/products"
	ctx := r.Context()

	productIDStr := chi.URLParam(r, "id")
	if productIDStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "Invalid product ID"))
		return
	}

	result := "success"
	defer func() {
		if err := s.services.staffLog.CreateLog(
			context.Background(),
			s.sessionManager.GetString(ctx, SessionStaffID),
			constants.ActionDelete,
			constants.ModuleProducts,
			result,
			nil,
		); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
	}()

	if err := s.services.product.Delete(ctx, productIDStr); err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("product_id", productIDStr),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError(page, "Failed to delete product"))
		return
	}

	result = fmt.Sprintf("success. ID '%s'", productIDStr)
	redirectHX(w, r, utils.URLWithSuccess(page, "Product deleted successfully"))
}

func (s *Server) adminSuperuserProductsEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Edit Page Handler]"
	const page = "/superuser/admin/products"
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	product, err := s.services.product.GetByIDForEdit(ctx, productID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("product_id", productID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

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

	var imageCDNURL string
	productImage, _ := s.dbRO.GetQueries().GetProductImageByProductID(ctx, product.ID)
	if productImage.Path != "" {
		imageCDNURL = s.GetCDNURL(productImage.Path)
	}

	formData := models.AdminProductEditForm{
		ProductID:   productID,
		Serial:      product.Serial,
		Name:        product.Name,
		Description: product.Description,
		BrandID:     s.encoder.Encode(product.BrandID),
		BrandName:   product.BrandName,
		Category:    product.Category,
		Subcategory: product.Subcategory,
		Price:       strconv.FormatInt(product.UnitPriceWithVat/100, 10),
		Status:      enums.ParseProductStatusToEnum(product.Status),
		ImageCDNURL: imageCDNURL,
		Specs: models.AdminProductSpecsForm{
			Colours:       product.Specs.Colours,
			Sizes:         product.Specs.Sizes,
			Segmentation:  product.Specs.Segmentation,
			PartNumber:    product.Specs.PartNumber,
			Power:         product.Specs.Power,
			Capacity:      product.Specs.Capacity,
			ScopeOfSupply: product.Specs.ScopeOfSupply,
			Weight:        product.Specs.Weight,
			WeightUnit:    product.Specs.WeightUnit,
		},
		Brands:        brands,
		Categories:    categories,
		VATPercentage: conf.Conf().Settings.VATPercentage,
	}

	if err := compadmin.AdminSuperuserProductsEditPage(formData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminSuperuserProductsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Update Handler]"
	const page = "/admin/superuser/products"
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		redirectHX(w, r, utils.URLWithError(page, "Invalid product ID"))
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	brandID := r.FormValue("brand_id")
	if brandID == "" {
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
	description := r.FormValue("description")
	priceStr := r.FormValue("price")
	statusStr := r.FormValue("status")
	if name == "" || description == "" || priceStr == "" || statusStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	status := enums.ParseProductStatusToEnum(statusStr)
	if status == enums.PRODUCT_STATUS_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, "Invalid status"))
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

	specs := services.ProductSpecsInput{
		Colours:       r.FormValue("spec_colours"),
		Sizes:         r.FormValue("spec_sizes"),
		Segmentation:  r.FormValue("spec_segmentation"),
		PartNumber:    r.FormValue("spec_part_number"),
		Power:         r.FormValue("spec_power"),
		Capacity:      r.FormValue("spec_capacity"),
		ScopeOfSupply: r.FormValue("spec_scope_of_supply"),
		Weight:        r.FormValue("spec_weight"),
		WeightUnit:    enums.ParseWeightUnitToEnum(r.FormValue("spec_weight_unit")).ToDB(),
	}

	var filename string
	var brandName string
	if conf.Conf().Test.LocalUploadImage || conf.Conf().IsProd() {
		file, header, err := r.FormFile("product_image")
		if err == nil {
			defer file.Close()

			brandName, err = s.services.brand.GetNameByID(ctx, brandID)
			if err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, "Brand not found"))
				return
			}

			filename = s.services.image.GenerateFilename(filepath.Ext(header.Filename), brandName, name)
			buf := bytes.Buffer{}
			if _, err := io.Copy(&buf, file); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, "Failed to read image"))
				return
			}

			contentType := header.Header.Get("Content-Type")
			if err := s.services.image.UploadProductImage(
				ctx,
				brandName,
				filename,
				&buf,
				contentType,
			); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, "Failed to upload image"))
				return
			}
		}
	}

	input := services.UpdateProductInput{
		ProductID:           productID,
		BrandID:             brandID,
		Category:            category,
		Subcategory:         subcategory,
		Name:                name,
		Description:         description,
		Specs:               specs,
		Status:              status.String(),
		ImagePath:           filename,
		UnitPriceWithoutVat: unitPriceWithoutVat,
		UnitPriceWithVat:    unitPriceWithVat,
	}

	result := "success"
	defer func() {
		if err := s.services.staffLog.CreateLog(
			context.Background(),
			s.sessionManager.GetString(ctx, SessionStaffID),
			constants.ActionUpdate,
			constants.ModuleProducts,
			result,
			nil,
		); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
	}()

	if err := s.services.product.Update(ctx, input); err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to update product"))
		return
	}

	if filename != "" && s.thumbnailJobRunner != nil {
		decodedProductID := s.encoder.Decode(productID)
		if err := s.thumbnailJobRunner.QueueThumbnailJob(ctx, jobs.ThumbnailJobParams{
			ProductID:  decodedProductID,
			Brand:      brandName,
			SourcePath: filename,
			Filename:   filepath.Base(filename),
		}); err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		}
	}

	result = fmt.Sprintf("success. ID '%s'", productID)
	redirectHX(w, r, utils.URLWithSuccess(page, "Product updated successfully"))
}
