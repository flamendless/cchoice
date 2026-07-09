package forms

import (
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/errs"
)

type AdminStaffChangePasswordForm struct {
	NewPassword     string `form:"new_password" validate:"required,ph_password"`
	ConfirmPassword string `form:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type AdminStaffProfileUpdateForm struct {
	FirstName  string `form:"first_name" validate:"required"`
	MiddleName string `form:"middle_name"`
	LastName   string `form:"last_name" validate:"required"`
	MobileNo   string `form:"mobile_no" validate:"required"`
	Birthdate  string `form:"birthdate" validate:"required"`
	DateHired  string `form:"date_hired" validate:"required"`
}

func (f *AdminStaffProfileUpdateForm) Normalize() {
	if f.MobileNo != "" && !strings.HasPrefix(f.MobileNo, constants.PHMobilePrefix) {
		f.MobileNo = constants.PHMobilePrefix + f.MobileNo
	}
}

func (f AdminStaffProfileUpdateForm) Validate() error {
	if !constants.ReMobileNumber.MatchString(f.MobileNo) {
		return errs.ErrValidationInvalidMobileNumber
	}
	return nil
}

type AdminStaffAttendanceDateQuery struct {
	Date string `form:"date"`
}

type AdminStaffAttendanceRowsQuery struct {
	DateSelector string `form:"date-selector"`
}

type AdminStaffTimeOffForm struct {
	Type        string `form:"type" validate:"required"`
	Description string `form:"description" validate:"required"`
	StartDate   string `form:"start-date" validate:"required"`
	EndDate     string `form:"end-date" validate:"required"`
}
