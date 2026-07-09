package server

import (
	"bytes"
	"cmp"
	"context"
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
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminSuperuserProductsSubcategoriesHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Subcategories Handler]"
	ctx := r.Context()

	var q forms.AdminProductsCategoryQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	category := q.Category

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

	var q forms.AdminProductsSerialQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}
	serial := q.Serial
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
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(errs.ErrInvalidParams)))
		return
	}

	var f forms.AdminProductCreateOrUpdateForm
	if err := httputil.BindMultipartForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductAllFieldsRequired.Error()))
		return
	}

	brandID := f.BrandID
	if brandID == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidBrand.Error()))
		return
	}

	category := f.Category
	subcategory := f.Subcategory
	if category == "" || subcategory == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductCategoryRequired.Error()))
		return
	}

	name := f.Name
	serial := f.Serial
	description := f.Description
	priceStr := f.Price
	if name == "" || serial == "" || description == "" || priceStr == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductAllFieldsRequired.Error()))
		return
	}

	isUnique, err := s.services.product.ValidateSerial(ctx, serial)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductSerialValidateFailed.Error()))
		return
	}

	if !isUnique {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductSerialExists.Error()))
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidPrice.Error()))
		return
	}

	vatPercentage, err := strconv.ParseFloat(conf.Conf().Settings.VATPercentage, 64)
	if err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
		vatPercentage = 0
	}
	unitPriceWithoutVat := int64(math.Round(price / (1 + vatPercentage/100)))
	unitPriceWithVat := int64(math.Round(price))

	specColours := f.SpecColours
	specSizes := f.SpecSizes
	specSegmentation := f.SpecSegmentation
	specPartNumber := f.SpecPartNumber
	specPower := f.SpecPower
	specCapacity := f.SpecCapacity
	specScopeOfSupply := f.SpecScopeOfSupply
	specWeight := f.SpecWeight
	specWeightUnit := f.SpecWeightUnit
	stocksInStr := f.StocksIn
	stocksQtyStr := f.StocksQty

	if specColours == "" || specSizes == "" || specSegmentation == "" ||
		specPartNumber == "" || specPower == "" || specCapacity == "" || specScopeOfSupply == "" ||
		specWeight == "" || specWeightUnit == "" || stocksInStr == "" || stocksQtyStr == "" {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err), zap.Any("form", r.Form))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductSpecsRequired.Error()))
		return
	}

	stocksIn := enums.ParseStocksInToEnum(stocksInStr)
	if stocksIn == enums.STOCKS_IN_UNDEFINED {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err), zap.String("stocks in", stocksInStr))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidStocksIn.Error()))
		return
	}

	stocksQty, err := strconv.ParseInt(stocksQtyStr, 10, 64)
	if err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err), zap.String("stocks qty", stocksQtyStr))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	salePriceStr := f.SalePrice
	saleStartDate := f.SaleStartDate
	saleEndDate := f.SaleEndDate

	var salePriceWithoutVat, salePriceWithVat int64
	if salePriceStr != "" {
		salePrice, parseErr := strconv.ParseFloat(salePriceStr, 64)
		if parseErr != nil || salePrice <= 0 {
			logs.LogCtx(ctx).Warn(logtag, zap.Error(parseErr), zap.String("sale price", salePriceStr))
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidSalePrice.Error()))
			return
		}

		if saleStartDate == "" || saleEndDate == "" {
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductSaleDatesRequired.Error()))
			return
		}

		salePriceWithoutVat = int64(math.Round(salePrice / (1 + vatPercentage/100)))
		salePriceWithVat = int64(math.Round(salePrice))
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
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductImageRequired.Error()))
			return
		}
		defer file.Close()

		brandName, err = s.services.brand.GetNameByID(ctx, brandID)
		if err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, errs.ErrBrandNotFound.Error()))
			return
		}

		filename = s.services.image.GenerateFilename(
			enums.IMAGE_PREFIX_PRODUCT_IMAGE,
			filepath.Ext(header.Filename),
			brandName,
			name,
		)
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, file); err != nil {
			result = err.Error()
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, errs.ErrFileRead.Error()))
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
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductImageUploadFailed.Error()))
			return
		}
	}

	externalLinks, err := parseExternalPlatformLinksFromForm(r)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	product, err := s.services.product.Create(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		services.CreateProductInput{
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
			SalePriceWithoutVat: salePriceWithoutVat,
			SalePriceWithVat:    salePriceWithVat,
			SaleStartDate:       saleStartDate,
			SaleEndDate:         saleEndDate,
			StocksIn:            stocksIn,
			Stocks:              stocksQty,
			ExternalLinks:       externalLinks,
		})
	if err != nil || product == nil {
		err = cmp.Or(err, errs.ErrServerProductNil)
		result = err.Error()
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
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
		brandsRes = []queries.GetBrandsForProductCreateRow{}
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

	brandsRes, err := requests.GetBrandsForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminBrandsCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		brandsRes = []queries.GetBrandsForProductCreateRow{}
	}

	brands := make([]models.AdminBrand, 0, len(brandsRes))
	for _, b := range brandsRes {
		brands = append(brands, models.AdminBrand{
			ID:   s.encoder.Encode(b.ID),
			Name: b.Name,
		})
	}

	var q forms.AdminProductsListQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	if err := compadmin.AdminSuperuserProductsListPage(
		"Products",
		brands,
		q.SearchSerial,
		q.Status,
	).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserProductsListTableHandler(w http.ResponseWriter, r *http.Request) {
	s.renderAdminProductsListTable(
		w, r,
		"/admin/superuser/products/table",
		"/admin/superuser/products",
		false,
		"/admin/superuser/products",
		"[Admin Superuser Products List Table Handler]",
	)
}

func (s *Server) adminSuperuserProductsUpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Update Status Handler]"
	const page = "/admin/superuser/products"
	ctx := r.Context()

	var p forms.AdminProductPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
		return
	}
	productIDStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
		return
	}

	var f forms.AdminProductStatusForm
	if err := httputil.BindForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	statusStr := f.Status
	if statusStr == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductStatusRequired.Error()))
		return
	}

	status := enums.ParseProductStatusToEnum(statusStr)
	if status == enums.PRODUCT_STATUS_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidStatus.Error()))
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
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductUpdateStatusFailed.Error()))
		return
	}

	result = fmt.Sprintf("success. ID '%s'", productIDStr)
	redirectHX(w, r, utils.URLWithSuccess(page, "Product status updated successfully"))
}

func (s *Server) adminSuperuserProductsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Delete Handler]"
	const page = "/admin/superuser/products"
	ctx := r.Context()

	var p forms.AdminProductPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
		return
	}
	productIDStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
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
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductDeleteFailed.Error()))
		return
	}

	result = fmt.Sprintf("success. ID '%s'", productIDStr)
	redirectHX(w, r, utils.URLWithSuccess(page, "Product deleted successfully"))
}

func (s *Server) adminSuperuserProductsEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Edit Page Handler]"
	const page = "/admin/superuser/products"
	ctx := r.Context()

	var p forms.AdminProductPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}
	productID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
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
		brandsRes = []queries.GetBrandsForProductCreateRow{}
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
	productImage, err := s.dbRO.GetQueries().GetProductImageByProductID(ctx, product.ID)
	if err != nil {
		logs.Log().Warn(page, zap.String("product id", productID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if productImage.Path != "" {
		imageCDNURL = s.GetCDNURL(productImage.Path)
	}

	inventory, err := s.services.productInventory.GetByProductID(ctx, productID)
	if err != nil || inventory == nil {
		err = cmp.Or(err, errs.ErrDBNil)
		logs.Log().Warn(page, zap.String("product id", productID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
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
			WeightUnit:    enums.ParseWeightUnitToEnum(product.Specs.WeightUnit),
		},
		Brands:        brands,
		Categories:    categories,
		VATPercentage: conf.Conf().Settings.VATPercentage,
		StocksIn:      inventory.StocksIn,
		Stocks:        strconv.FormatInt(inventory.Stocks, 10),
		UpdateURL:     "/admin/superuser/products/" + productID,
		ListPageURL:   "/admin/superuser/products",
		ExternalLinks: toAdminProductExternalLinks(product.ExternalLinks),
	}
	if product.SalePriceWithVat > 0 {
		formData.SalePrice = strconv.FormatFloat(float64(product.SalePriceWithVat)/100, 'f', -1, 64)
		formData.SaleStartDate = product.SaleStartDate
		formData.SaleEndDate = product.SaleEndDate
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

	var p forms.AdminProductPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
		return
	}
	productID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(errs.ErrInvalidParams)))
		return
	}

	var f forms.AdminProductCreateOrUpdateForm
	if err := httputil.BindMultipartForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductAllFieldsRequired.Error()))
		return
	}
	brandID := f.BrandID
	if brandID == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidBrand.Error()))
		return
	}

	category := f.Category
	subcategory := f.Subcategory
	if category == "" || subcategory == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductCategoryRequired.Error()))
		return
	}

	name := f.Name
	description := f.Description
	priceStr := f.Price
	statusStr := f.Status
	if name == "" || description == "" || priceStr == "" || statusStr == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductAllFieldsRequired.Error()))
		return
	}

	status := enums.ParseProductStatusToEnum(statusStr)
	if status == enums.PRODUCT_STATUS_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidStatus.Error()))
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidPrice.Error()))
		return
	}

	vatPercentage, err := strconv.ParseFloat(conf.Conf().Settings.VATPercentage, 64)
	if err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
		vatPercentage = 0
	}
	unitPriceWithoutVat := int64(math.Round(price / (1 + vatPercentage/100)))
	unitPriceWithVat := int64(math.Round(price))

	salePriceStr := f.SalePrice
	saleStartDate := f.SaleStartDate
	saleEndDate := f.SaleEndDate

	var salePriceWithoutVat, salePriceWithVat int64
	if salePriceStr != "" {
		salePrice, parseErr := strconv.ParseFloat(salePriceStr, 64)
		if parseErr != nil || salePrice <= 0 {
			logs.LogCtx(ctx).Warn(logtag, zap.Error(parseErr), zap.String("sale price", salePriceStr))
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidSalePrice.Error()))
			return
		}

		if saleStartDate == "" || saleEndDate == "" {
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductSaleDatesRequired.Error()))
			return
		}

		salePriceWithoutVat = int64(math.Round(salePrice / (1 + vatPercentage/100)))
		salePriceWithVat = int64(math.Round(salePrice))
	}

	specs := services.ProductSpecsInput{
		Colours:       f.SpecColours,
		Sizes:         f.SpecSizes,
		Segmentation:  f.SpecSegmentation,
		PartNumber:    f.SpecPartNumber,
		Power:         f.SpecPower,
		Capacity:      f.SpecCapacity,
		ScopeOfSupply: f.SpecScopeOfSupply,
		Weight:        f.SpecWeight,
		WeightUnit:    enums.ParseWeightUnitToEnum(f.SpecWeightUnit).ToDB(),
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
				redirectHX(w, r, utils.URLWithError(page, errs.ErrBrandNotFound.Error()))
				return
			}

			filename = s.services.image.GenerateFilename(
				enums.IMAGE_PREFIX_PRODUCT_IMAGE,
				filepath.Ext(header.Filename),
				brandName,
				name,
			)
			buf := bytes.Buffer{}
			if _, err := io.Copy(&buf, file); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, errs.ErrFileRead.Error()))
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
				redirectHX(w, r, utils.URLWithError(page, errs.ErrProductImageUploadFailed.Error()))
				return
			}
		}
	}

	qty, err := strconv.ParseInt(f.StocksQty, 10, 64)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	externalLinks, err := parseExternalPlatformLinksFromForm(r)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
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
		SalePriceWithoutVat: salePriceWithoutVat,
		SalePriceWithVat:    salePriceWithVat,
		SaleStartDate:       saleStartDate,
		SaleEndDate:         saleEndDate,
		StocksIn:            enums.MustParseStocksInToEnum(f.StocksIn),
		Stocks:              qty,
		ExternalLinks:       externalLinks,
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

	if err := s.services.product.Update(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		input,
	); err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
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

func (s *Server) adminStaffProductsListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Products List Page Handler]"
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
		brandsRes = []queries.GetBrandsForProductCreateRow{}
	}

	brands := make([]models.AdminBrand, 0, len(brandsRes))
	for _, b := range brandsRes {
		brands = append(brands, models.AdminBrand{
			ID:   s.encoder.Encode(b.ID),
			Name: b.Name,
		})
	}

	if err := compadmin.AdminStaffProductsListPage(brands).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffProductsListTableHandler(w http.ResponseWriter, r *http.Request) {
	s.renderAdminProductsListTable(
		w, r,
		"/admin/products/table",
		"/admin/products",
		true,
		"/admin/products",
		"[Admin Staff Products List Table Handler]",
	)
}

func (s *Server) adminStaffProductsEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Products Edit Page Handler]"
	const page = "/admin/products"
	ctx := r.Context()

	var p forms.AdminProductPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}
	productID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	product, err := s.services.product.GetByIDForEdit(ctx, productID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("product_id", productID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if product.Status != enums.PRODUCT_STATUS_DRAFT.String() {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductDraftOnly.Error()))
		return
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
	productImage, err := s.dbRO.GetQueries().GetProductImageByProductID(ctx, product.ID)
	if err != nil {
		logs.Log().Warn(logtag, zap.String("product id", productID), zap.Error(err))
	} else if productImage.Path != "" {
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
			WeightUnit:    enums.ParseWeightUnitToEnum(product.Specs.WeightUnit),
		},
		Categories:     categories,
		VATPercentage:  conf.Conf().Settings.VATPercentage,
		StocksIn:       product.StocksIn,
		Stocks:         strconv.FormatInt(product.Stocks, 10),
		StatusReadOnly: true,
		DraftPriceOnly: true,
		UpdateURL:      "/admin/products/" + productID,
		ListPageURL:    page,
		ExternalLinks:  toAdminProductExternalLinks(product.ExternalLinks),
	}
	if product.SalePriceWithVat > 0 {
		formData.SalePrice = strconv.FormatFloat(float64(product.SalePriceWithVat)/100, 'f', -1, 64)
		formData.SaleStartDate = product.SaleStartDate
		formData.SaleEndDate = product.SaleEndDate
	}

	if err := compadmin.AdminSuperuserProductsEditPage(formData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminStaffProductsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Products Update Handler]"
	const page = "/admin/products"
	ctx := r.Context()

	var p forms.AdminProductPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
		return
	}
	productID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidID.Error()))
		return
	}

	product, err := s.services.product.GetByIDForEdit(ctx, productID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("product_id", productID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if product.Status != enums.PRODUCT_STATUS_DRAFT.String() {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductDraftOnly.Error()))
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(errs.ErrInvalidParams)))
		return
	}

	var f forms.AdminProductCreateOrUpdateForm
	if err := httputil.BindMultipartForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductAllFieldsRequired.Error()))
		return
	}
	category := f.Category
	subcategory := f.Subcategory
	if category == "" || subcategory == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductCategoryRequired.Error()))
		return
	}

	name := f.Name
	description := f.Description
	priceStr := f.Price
	if name == "" || description == "" || priceStr == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductAllFieldsRequired.Error()))
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidPrice.Error()))
		return
	}

	vatPercentage, err := strconv.ParseFloat(conf.Conf().Settings.VATPercentage, 64)
	if err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
		vatPercentage = 0
	}
	unitPriceWithoutVat := int64(math.Round(price / (1 + vatPercentage/100)))
	unitPriceWithVat := int64(math.Round(price))

	salePriceStr := f.SalePrice
	saleStartDate := f.SaleStartDate
	saleEndDate := f.SaleEndDate

	var salePriceWithoutVat, salePriceWithVat int64
	if salePriceStr != "" {
		salePrice, parseErr := strconv.ParseFloat(salePriceStr, 64)
		if parseErr != nil || salePrice <= 0 {
			logs.LogCtx(ctx).Warn(logtag, zap.Error(parseErr), zap.String("sale price", salePriceStr))
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidSalePrice.Error()))
			return
		}

		if saleStartDate == "" || saleEndDate == "" {
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductSaleDatesRequired.Error()))
			return
		}

		salePriceWithoutVat = int64(math.Round(salePrice / (1 + vatPercentage/100)))
		salePriceWithVat = int64(math.Round(salePrice))
	}

	specs := services.ProductSpecsInput{
		Colours:       f.SpecColours,
		Sizes:         f.SpecSizes,
		Segmentation:  f.SpecSegmentation,
		PartNumber:    f.SpecPartNumber,
		Power:         f.SpecPower,
		Capacity:      f.SpecCapacity,
		ScopeOfSupply: f.SpecScopeOfSupply,
		Weight:        f.SpecWeight,
		WeightUnit:    enums.ParseWeightUnitToEnum(f.SpecWeightUnit).ToDB(),
	}

	if specs.Colours == "" || specs.Sizes == "" || specs.Segmentation == "" ||
		specs.PartNumber == "" || specs.Power == "" || specs.Capacity == "" ||
		specs.ScopeOfSupply == "" || specs.Weight == "" || specs.WeightUnit == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductSpecsRequired.Error()))
		return
	}

	brandID := s.encoder.Encode(product.BrandID)
	var filename string
	var brandName string
	if conf.Conf().Test.LocalUploadImage || conf.Conf().IsProd() {
		file, header, err := r.FormFile("product_image")
		if err == nil {
			defer file.Close()

			brandName, err = s.services.brand.GetNameByID(ctx, brandID)
			if err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, errs.ErrBrandNotFound.Error()))
				return
			}

			filename = s.services.image.GenerateFilename(
				enums.IMAGE_PREFIX_PRODUCT_IMAGE,
				filepath.Ext(header.Filename),
				brandName,
				name,
			)
			buf := bytes.Buffer{}
			if _, err := io.Copy(&buf, file); err != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				redirectHX(w, r, utils.URLWithError(page, errs.ErrFileRead.Error()))
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
				redirectHX(w, r, utils.URLWithError(page, errs.ErrProductImageUploadFailed.Error()))
				return
			}
		}
	}

	externalLinks, err := parseExternalPlatformLinksFromForm(r)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	input := services.UpdateProductInput{
		ProductID:           productID,
		BrandID:             brandID,
		Category:            category,
		Subcategory:         subcategory,
		Name:                name,
		Description:         description,
		Specs:               specs,
		Status:              product.Status,
		ImagePath:           filename,
		UnitPriceWithoutVat: unitPriceWithoutVat,
		UnitPriceWithVat:    unitPriceWithVat,
		SalePriceWithoutVat: salePriceWithoutVat,
		SalePriceWithVat:    salePriceWithVat,
		SaleStartDate:       saleStartDate,
		SaleEndDate:         saleEndDate,
		StocksIn:            product.StocksIn,
		Stocks:              product.Stocks,
		ExternalLinks:       externalLinks,
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

	if err := s.services.product.Update(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		input,
	); err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
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

const productListFilterInclude = "[name='search_serial'],[name='search_brand'],[name='status']"

func parseAdminProductListPage(r *http.Request) int {
	var q forms.AdminProductsListQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		return 1
	}
	return httputil.PageOrDefault(q.Page, 1)
}

func (s *Server) renderAdminProductsListTable(
	w http.ResponseWriter,
	r *http.Request,
	tableURL string,
	basePath string,
	editOnly bool,
	errorPage string,
	logtag string,
) {
	ctx := r.Context()

	var q forms.AdminProductsListQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(errorPage, httputil.ErrorMessage(err)))
		return
	}
	searchSerial := q.SearchSerial
	searchBrand := q.SearchBrand

	statusStr := q.Status
	status := enums.ParseProductStatusToEnum(statusStr)
	if statusStr != "" && status == enums.PRODUCT_STATUS_UNDEFINED {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("search serial", searchSerial),
			zap.String("search brand", searchBrand),
			zap.String("status", statusStr),
			zap.Error(errs.ErrEnumInvalid),
		)
		redirectHX(w, r, utils.URLWithError(errorPage, errs.ErrEnumInvalid.Error()))
		return
	}

	page := parseAdminProductListPage(r)

	productList, totalCount, page, err := s.services.product.GetForListingAdminPaginated(
		ctx,
		searchSerial,
		searchBrand,
		status,
		page,
		constants.DefaultAdminTablePageSize,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		productList = []models.AdminProductListItem{}
		totalCount = 0
	}

	pagination := models.TablePagination{
		Page:          page,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL(tableURL),
		Include:       productListFilterInclude,
		ContentTarget: "#products-table-content",
	}

	if err := compadmin.AdminSuperuserProductsListTableContent(productList, editOnly, basePath, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("search", searchSerial),
			zap.String("status", statusStr),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError(errorPage, err.Error()))
		return
	}
}

func parseExternalPlatformLinksFromForm(r *http.Request) ([]services.ExternalPlatformLinkInput, error) {
	platforms := r.Form["external_platform[]"]
	if len(platforms) == 0 {
		platforms = r.Form["external_platform"]
	}

	urls := r.Form["external_url[]"]
	if len(urls) == 0 {
		urls = r.Form["external_url"]
	}

	maxLen := max(len(platforms), len(urls))
	links := make([]services.ExternalPlatformLinkInput, 0, maxLen)
	seen := make(map[string]struct{}, maxLen)

	for i := range maxLen {
		platform := ""
		url := ""
		if i < len(platforms) {
			platform = strings.TrimSpace(platforms[i])
		}
		if i < len(urls) {
			url = strings.TrimSpace(urls[i])
		}

		if platform == "" && url == "" {
			continue
		}

		if platform == "" || url == "" {
			return nil, errs.ErrInvalidExternalPlatformLink
		}

		platformEnum := enums.ParseExternalPlatformToEnum(platform)
		if platformEnum == enums.EXTERNAL_PLATFORM_UNDEFINED {
			return nil, errs.ErrInvalidExternalPlatformLink
		}

		platformKey := platformEnum.String()
		if _, exists := seen[platformKey]; exists {
			return nil, errs.ErrInvalidExternalPlatformLink
		}
		seen[platformKey] = struct{}{}

		if err := utils.ValidateExternalURL(url); err != nil {
			return nil, errs.ErrInvalidExternalPlatformLink
		}

		links = append(links, services.ExternalPlatformLinkInput{
			Platform: platformKey,
			URL:      url,
		})
	}

	return links, nil
}

func toAdminProductExternalLinks(links []services.ExternalPlatformLinkInput) []models.AdminProductExternalLink {
	result := make([]models.AdminProductExternalLink, 0, len(links))
	for _, link := range links {
		result = append(result, models.AdminProductExternalLink{
			Platform: enums.ParseExternalPlatformToEnum(link.Platform),
			URL:      link.URL,
		})
	}
	return result
}
