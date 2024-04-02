package utils

import (
	"errors"

	v "github.com/cohesivestack/valgo"
)

func StringsToError(errs []string) error {
	var err error
	for i := 0; i < len(errs); i++ {
		err = errors.Join(errors.New(errs[i]))
	}
	return err
}

func ValidateNotBlank(field string, key string) error {
	val := v.Check(
		v.String(field, key).Not().Blank(),
	)

	if !val.Valid() {
		errs := val.Errors()[key]
		return StringsToError(errs.Messages())
	}

	return nil
}
