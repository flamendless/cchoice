package server

import (
	"net/http"
	"strconv"
	"strings"

	compcustomer "cchoice/cmd/web/components/customers"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCustomerID       = "customer_id"
	SessionCustomerAccessID = "customer_access_id"
)

func AddCustomerHandlers(s *Server, r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(s.rateLimiter.Middleware)
		r.Post("/customer/login", s.customerLoginHandler)
		r.Post("/customer/register", s.customerRegisterHandler)
	})

	r.Get("/customer", s.customerLoginPageHandler)
	r.Get("/customer/register", s.customerRegisterPageHandler)
	r.With(s.requireCustomerAuth).Post("/customer/logout", s.customerLogoutHandler)
	r.With(s.requireCustomerAuth).Get("/customer/portal", s.customerPortalHandler)

	r.With(s.requireCustomerAuth).Get("/customer/quotation", s.customerQuotationPageHandler)
	r.With(s.requireCustomerAuth).Post("/customer/quotation/product/{productID}/add", s.customerQuotationAddToHandler)
	r.With(s.requireCustomerAuth).Delete("/customer/quotation/line/{lineID}/remove", s.customerQuotationRemoveFromDraftHandler)
	r.With(s.requireCustomerAuth).Post("/customer/quotation/submit", s.customerQuotationSubmitDraftHandler)

	r.With(s.requireCustomerAuth).Get("/customer/profile", s.customerProfileHandler)
	r.With(s.requireCustomerAuth).Get("/customer/profile/edit", s.customerProfileEditFormHandler)
	r.With(s.requireCustomerAuth).Patch("/customer/profile", s.customerProfileUpdateHandler)
	r.With(s.requireCustomerAuth).Post("/customer/change-password", s.customerChangePasswordHandler)
	r.With(s.requireCustomerAuth).Post("/customer/verify/send", s.customerVerifySendHandler)
	r.With(s.requireCustomerAuth).Post("/customer/verify", s.customerVerifyHandler)
}

func (s *Server) customerLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Login Page Handler]"
	ctx := r.Context()

	if s.sessionManager.GetString(ctx, SessionCustomerID) != "" {
		redirectHX(w, r, utils.URL("/customer/portal"))
		return
	}

	if err := compcustomer.CustomerLoginPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerLoginHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Login Handler]"
	const page = "/customer"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	email := r.PostFormValue("email")
	if !constants.ReEmail.MatchString(email) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid email or password format"))
		return
	}

	password := r.PostFormValue("password")
	if !constants.RePassword.MatchString(password) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid email or password format"))
		return
	}

	customer, err := s.services.customer.Login(ctx, email, password)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid email or password"))
		return
	}

	s.sessionManager.Put(ctx, SessionCustomerID, s.encoder.Encode(customer.ID))
	s.sessionManager.Put(ctx, SessionCustomerAccessID, 0)
	redirectHX(w, r, utils.URL("/customer/portal"))
}

func (s *Server) customerLogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.sessionManager.Remove(ctx, SessionCustomerID)
	s.sessionManager.Remove(ctx, SessionCustomerAccessID)
	redirectHX(w, r, utils.URL("/customer"))
}

func (s *Server) customerRegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Register Page Handler]"
	ctx := r.Context()

	if err := compcustomer.CustomerRegisterPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerRegisterHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Register Handler]"
	const page = "/customer/register"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	firstName := r.PostFormValue("first_name")
	middleName := r.PostFormValue("middle_name")
	lastName := r.PostFormValue("last_name")
	birthdate := r.PostFormValue("birthdate")
	sex := r.PostFormValue("sex")
	email := r.PostFormValue("email")
	mobileNo := r.PostFormValue("mobile_no")
	password := r.PostFormValue("password")
	confirmPassword := r.PostFormValue("confirm_password")
	customerType := r.PostFormValue("customer_type")
	companyName := r.PostFormValue("company_name")

	if password != confirmPassword {
		redirectHX(w, r, utils.URLWithError(page, "Passwords must match"))
		return
	}

	if firstName == "" || lastName == "" || birthdate == "" || sex == "" || email == "" || mobileNo == "" || password == "" || customerType == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	if !constants.ReEmail.MatchString(email) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid email format"))
		return
	}

	if !strings.HasPrefix(mobileNo, constants.PHMobilePrefix) {
		mobileNo = constants.PHMobilePrefix + mobileNo
	}

	if !constants.ReMobileNumber.MatchString(mobileNo) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid mobile number format"))
		return
	}

	if !constants.RePassword.MatchString(password) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid password format"))
		return
	}

	customerTypeEnum := enums.MustParseCustomerTypeToEnum(customerType)
	if customerTypeEnum == enums.CUSTOMER_TYPE_COMPANY && companyName == "" {
		redirectHX(w, r, utils.URLWithError(page, "Company name is required for company accounts"))
		return
	}

	if _, err := s.services.customer.Register(ctx, services.RegisterCustomerParams{
		FirstName:    firstName,
		MiddleName:   middleName,
		LastName:     lastName,
		Birthdate:    birthdate,
		Sex:          sex,
		Email:        email,
		MobileNo:     mobileNo,
		Password:     password,
		CustomerType: customerTypeEnum,
		CompanyName:  companyName,
	}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess("/customer", "Registration successful! Please log in."))
}

func (s *Server) customerPortalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Portal Handler]"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	profile, err := s.services.customer.BuildProfile(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/customer", "Unable to load profile"))
		return
	}

	if err := compcustomer.CustomerPortalPage(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerProfileHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Profile Handler]"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	profile, err := s.services.customer.BuildProfile(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/customer/portal", "Unable to load profile"))
		return
	}

	if err := compcustomer.CustomerProfilePage(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerProfileEditFormHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Profile Edit Form Handler]"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	profile, err := s.services.customer.BuildProfile(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/customer/portal", "Unable to load profile"))
		return
	}

	if err := compcustomer.CustomerProfileEditPage(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Profile Update Handler]"
	const page = "/customer/profile"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	firstName := r.PostFormValue("first_name")
	middleName := r.PostFormValue("middle_name")
	lastName := r.PostFormValue("last_name")
	mobileNo := r.PostFormValue("mobile_no")
	birthdate := r.PostFormValue("birthdate")
	sex := r.PostFormValue("sex")

	if firstName == "" || lastName == "" || birthdate == "" || sex == "" || mobileNo == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	if !strings.HasPrefix(mobileNo, constants.PHMobilePrefix) {
		mobileNo = constants.PHMobilePrefix + mobileNo
	}

	if !constants.ReMobileNumber.MatchString(mobileNo) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid mobile number format"))
		return
	}

	err := s.services.customer.UpdateProfile(ctx, services.UpdateCustomerProfileParams{
		ID:         customerIDStr,
		FirstName:  firstName,
		MiddleName: middleName,
		LastName:   lastName,
		MobileNo:   mobileNo,
		Birthdate:  birthdate,
		Sex:        sex,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to update profile"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Profile updated successfully"))
}

func (s *Server) customerChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Change Password Handler]"
	const page = "/customer/page"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	currentPassword := r.PostFormValue("current_password")
	newPassword := r.PostFormValue("new_password")
	confirmPassword := r.PostFormValue("confirm_password")

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		redirectHX(w, r, utils.URLWithError(page, "All password fields are required"))
		return
	}

	if newPassword != confirmPassword {
		redirectHX(w, r, utils.URLWithError(page, "New passwords do not match"))
		return
	}

	if !constants.RePassword.MatchString(newPassword) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid password format"))
		return
	}

	customer, err := s.services.customer.GetByID(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Unable to verify current password"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(currentPassword)); err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Current password is incorrect"))
		return
	}

	err = s.services.customer.UpdatePassword(ctx, customerIDStr, newPassword)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to update password"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Password changed successfully"))
}

func (s *Server) customerVerifySendHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Verify Send Handler]"
	const page = "/customer/profile"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	customer, err := s.services.customer.GetByID(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Unable to send verification code"))
		return
	}

	if err := s.services.customerOTP.GenerateAndSendVerificationCode(ctx, services.GenerateOTPParams{
		CustomerID: customerIDStr,
		Email:      customer.Email,
	}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Verification code sent! Please check your email."))
}

func (s *Server) customerVerifyHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Verify Handler]"
	const page = "/customer/profile"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	otpCode := r.PostFormValue("otp_code")
	if otpCode == "" {
		redirectHX(w, r, utils.URLWithError(page, "Verification code is required"))
		return
	}

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	valid, err := s.services.customerOTP.VerifyCode(ctx, customerIDStr, otpCode)
	if err != nil || !valid {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid or expired verification code"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Email verified successfully!"))
}

func (s *Server) customerQuotationPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Portal Page Handler]"
	const page = "/customer"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	if _, err := s.services.customer.BuildProfile(ctx, customerIDStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	quotation, err := s.services.quotation.GetOrCreateActive(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	quotationID := s.encoder.Encode(quotation.ID)

	draftLines, err := s.services.quotation.GetLines(ctx, quotationID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	summary, err := s.services.quotation.GetSummary(ctx, quotationID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	products, err := s.services.product.ListForQuotations(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	brands, err := s.services.brand.GetAllActive(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
	activeBrands := make([]models.Brand, 0, len(brands))
	for _, brand := range brands {
		activeBrands = append(activeBrands, models.Brand{
			ID:     s.encoder.Encode(brand.ID),
			Name:   brand.Name,
			Status: brand.Status,
		})
	}

	categories, err := s.services.product.GetAllCategoryNames(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	subcategories, err := s.services.product.GetAllSubcategoryNames(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	draftLinesModel := make([]models.QuotationLine, 0, len(draftLines))
	for _, line := range draftLines {
		orig := utils.NewMoney(line.OriginalPriceSnapshot.Int64*line.Quantity, line.Currency)
		sale := utils.NewMoney(line.SalePriceSnapshot.Int64*line.Quantity, line.Currency)
		td, _ := orig.Subtract(sale)

		draftLinesModel = append(draftLinesModel, models.QuotationLine{
			ID:            s.encoder.Encode(line.ID),
			ProductSerial: line.ProductSerial,
			ProductName:   line.ProductName,
			BrandName:     line.BrandName,
			Quantity:      line.Quantity,
			TotalPrice:    orig.Display(),
			TotalDiscount: td.Display(),
		})
	}

	summaryModel := models.QuotationSummary{}
	if summary.TotalItems > 0 {
		summaryModel.TotalItems = summary.TotalItems

		var orig, discount int64
		if originalPrice, ok := summary.TotalOriginalPrice.(int64); ok {
			orig = originalPrice
		}
		if discounts, ok := summary.TotalSalePrice.(int64); ok {
			discount = discounts
		}

		origM := utils.NewMoney(orig, summary.Currency)
		discountM := utils.NewMoney(discount, summary.Currency)
		total, _ := origM.Subtract(discountM)
		summaryModel.TotalPrice = origM.Display()
		summaryModel.TotalDiscounts = discountM.Display()
		summaryModel.Total = total.Display()
	}

	productsModel := make([]models.QuotationProduct, 0, len(products))
	for _, p := range products {
		origPrice, discountedPrice, discountPercentage := utils.GetOrigAndDiscounted(
			p.IsOnSale,
			p.UnitPriceWithVat,
			p.UnitPriceWithVatCurrency,
			p.SalePriceWithVat,
			p.SalePriceWithVatCurrency,
		)

		productsModel = append(productsModel, models.QuotationProduct{
			ID:                 s.encoder.Encode(p.ID),
			Slug:               p.Slug.String,
			Serial:             p.Serial,
			Name:               p.Name,
			BrandName:          p.BrandName,
			Category:           p.Category.String,
			Subcategory:        p.Subcategory.String,
			DiscountPercentage: discountPercentage,
			OrigPriceDisplay:   origPrice.Display(),
			PriceDisplay:       discountedPrice.Display(),
			Quantity:           1,
		})
	}

	if err := compcustomer.QuotationPage(
		draftLinesModel,
		summaryModel,
		productsModel,
		activeBrands,
		categories,
		subcategories,
		"",
		"",
		"",
		"",
	).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerQuotationAddToHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Portal Add To Handler]"
	const page = "/customer/quotation"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	productID := chi.URLParam(r, "productID")
	var qty int64 = 1
	if qtyStr := r.FormValue("quantity"); qtyStr != "" {
		n, err := strconv.ParseInt(qtyStr, 10, 64)
		if err == nil {
			qty = n
		}
	}

	if err := s.services.quotation.AddLineToQuotation(ctx, customerIDStr, productID, qty); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URL(page))
}

func (s *Server) customerQuotationRemoveFromDraftHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Portal Remove From Quotation Handler]"
	const page = "/customer/quotation"
	ctx := r.Context()

	lineID := chi.URLParam(r, "lineID")

	if err := s.services.quotation.RemoveLine(ctx, lineID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URL(page))
}

func (s *Server) customerQuotationSubmitDraftHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Portal Submit Quotation Handler]"
	const page = "/customer/quotation"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	quotation, err := s.services.quotation.GetOrCreateActive(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if err := s.services.quotation.SubmitForReview(ctx, s.encoder.Encode(quotation.ID)); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Quotation submitted for review!"))
}
