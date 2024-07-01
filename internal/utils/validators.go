package utils

import (
	"errors"
	"fmt"

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
	fmt.Println(username, val.Valid())
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
