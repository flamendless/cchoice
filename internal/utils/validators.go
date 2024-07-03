package utils

import (
	pb "cchoice/proto"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	v "github.com/cohesivestack/valgo"
)

func ValidateNotBlank(field string, key string) error {
	val := v.Check(
		v.String(field, key).Not().Blank(),
	)

	if !val.Valid() {
		errs := val.Errors()[key]
		errMsg := fmt.Sprintf(
			"%s - %s",
			errs.Name(),
			errs.Messages(),
		)
		return errors.New(errMsg)
	}

	return nil
}

func ValidateUsername(username string) error {
	val := v.Check(
		v.String(username, "username").Not().Blank().OfLengthBetween(8, 32),
	)
	if !val.Valid() {
		errs := val.Errors()["username"]
		errMsg := fmt.Sprintf(
			"%s - %s",
			errs.Name(),
			errs.Messages(),
		)
		return errors.New(errMsg)
	}
	return nil
}

func ValidatePW(pw string) error {
	val := v.Check(
		v.String(pw, "password").Not().Blank().OfLengthBetween(8, 32),
	)
	if !val.Valid() {
		errs := val.Errors()["password"]
		errMsg := fmt.Sprintf(
			"%s - %s",
			errs.Name(),
			errs.Messages(),
		)
		return errors.New(errMsg)
	}
	return nil
}

func ValidateUserReg(data *pb.RegisterRequest) error {
	val := v.Check(
		v.String(data.FirstName, "first name").Not().Blank(),
		v.String(data.MiddleName, "middle name").Not().Blank(),
		v.String(data.LastName, "last name").Not().Blank(),
		v.String(data.Email, "email").Not().Blank(),
		v.String(data.Password, "password").Not().Blank().OfLengthBetween(8, 32),
		v.String(data.MobileNo, "mobile number").Not().Blank().OfLength(13),
	)

	if !val.Valid() {
		errs := val.Errors()
		errMsgs := make([]error, 0, len(errs))
		for _, err := range errs {
			errMsg := fmt.Errorf(
				"%s - %s",
				err.Name(),
				err.Messages(),
			)
			errMsgs = append(errMsgs, errMsg)
		}
		return errors.Join(errMsgs...)
	}

	_, err := mail.ParseAddress(data.Email)
	if err != nil {
		return err
	}

	validMobile := strings.HasPrefix(data.MobileNo, "+639")
	if !validMobile {
		return errors.New("Invalid mobile number format")
	}

	return nil
}
