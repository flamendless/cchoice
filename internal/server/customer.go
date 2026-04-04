package server

import (
	"database/sql"
	"net/http"

	compcustomer "cchoice/cmd/web/components/customers"
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
	r.Get("/customer", s.customerLoginPageHandler)
	r.Post("/customer/login", s.customerLoginHandler)
	r.Get("/customer/register", s.customerRegisterPageHandler)
	r.Post("/customer/register", s.customerRegisterHandler)
	r.With(s.requireCustomerAuth).Post("/customer/logout", s.customerLogoutHandler)
	r.With(s.requireCustomerAuth).Get("/customer/portal", s.customerPortalHandler)
	r.With(s.requireCustomerAuth).Get("/customer/profile", s.customerProfileHandler)
	r.With(s.requireCustomerAuth).Get("/customer/profile/edit", s.customerProfileEditFormHandler)
	r.With(s.requireCustomerAuth).Patch("/customer/profile", s.customerProfileUpdateHandler)
	r.With(s.requireCustomerAuth).Post("/customer/change-password", s.customerChangePasswordHandler)
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
	password := r.PostFormValue("password")

	if !constants.ReEmail.MatchString(email) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid email or password format"))
		return
	}

	if !constants.RePassword.MatchString(password) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid email or password format"))
		return
	}

	customer, err := s.services.customer.Login(ctx, email, password)
	if err != nil {
		if err == sql.ErrNoRows {
			redirectHX(w, r, utils.URLWithError(page, "Invalid email or password"))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
