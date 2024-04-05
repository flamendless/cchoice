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
