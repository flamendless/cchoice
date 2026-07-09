package forms

import (
	"errors"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
)

type CustomerRegisterForm struct {
	FirstName       string `form:"first_name" validate:"required"`
	MiddleName      string `form:"middle_name"`
	LastName        string `form:"last_name" validate:"required"`
	Birthdate       string `form:"birthdate" validate:"required"`
	Sex             string `form:"sex" validate:"required"`
	Email           string `form:"email" validate:"required,ph_email"`
	MobileNo        string `form:"mobile_no" validate:"required"`
	Password        string `form:"password" validate:"required,ph_password"`
	ConfirmPassword string `form:"confirm_password" validate:"required,eqfield=Password"`
	CustomerType    string `form:"customer_type" validate:"required"`
	CompanyName     string `form:"company_name"`
}

func (f *CustomerRegisterForm) Normalize() {
	if f.MobileNo != "" && !strings.HasPrefix(f.MobileNo, constants.PHMobilePrefix) {
		f.MobileNo = constants.PHMobilePrefix + f.MobileNo
	}
}

func (f CustomerRegisterForm) Validate() error {
	if !constants.ReMobileNumber.MatchString(f.MobileNo) {
		return errs.ErrValidationInvalidMobileNumber
	}
	if enums.ParseCustomerTypeToEnum(f.CustomerType) == enums.CUSTOMER_TYPE_COMPANY && f.CompanyName == "" {
		return errors.New("company name is required for company accounts")
	}
	return nil
}

type CustomerLoginForm struct {
	Email    string `form:"email" validate:"required,ph_email"`
	Password string `form:"password" validate:"required,ph_password"`
}

type CustomerProfileUpdateForm struct {
	FirstName  string `form:"first_name" validate:"required"`
	MiddleName string `form:"middle_name"`
	LastName   string `form:"last_name" validate:"required"`
	Birthdate  string `form:"birthdate" validate:"required"`
	Sex        string `form:"sex" validate:"required"`
	MobileNo   string `form:"mobile_no" validate:"required"`
}

func (f *CustomerProfileUpdateForm) Normalize() {
	if f.MobileNo != "" && !strings.HasPrefix(f.MobileNo, constants.PHMobilePrefix) {
		f.MobileNo = constants.PHMobilePrefix + f.MobileNo
	}
}

func (f CustomerProfileUpdateForm) Validate() error {
	if !constants.ReMobileNumber.MatchString(f.MobileNo) {
		return errs.ErrValidationInvalidMobileNumber
	}
	return nil
}

type CustomerChangePasswordForm struct {
	CurrentPassword string `form:"current_password" validate:"required"`
	NewPassword     string `form:"new_password" validate:"required,ph_password"`
	ConfirmPassword string `form:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type CustomerVerifyForm struct {
	OTPCode string `form:"otp_code" validate:"required"`
}

type CustomerQuotationAddForm struct {
	Quantity int `form:"quantity" validate:"omitempty,min=1"`
}

type CustomerQuotationPath struct {
	ProductID string `param:"productID" validate:"required"`
}

type CustomerQuotationLinePath struct {
	LineID string `param:"lineID" validate:"required"`
}

type CustomerOrderPath struct {
	ID string `param:"id" validate:"required"`
}

type CustomerOrdersListQuery struct {
	SearchOrderRef string `form:"search_order_ref"`
	Page           int    `form:"page"`
}

type CustomerQuotationsListQuery struct {
	Page int `form:"page"`
}

type CustomerQuotationDetailPath struct {
	ID string `param:"id" validate:"required"`
}
