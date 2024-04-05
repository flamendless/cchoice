package utils

import "errors"

func StringsToError(errs []string) error {
	var err error
	for i := 0; i < len(errs); i++ {
		err = errors.Join(errors.New(errs[i]))
	}
	return err
}

func JoinErrors(errs []error) string {
	var temp error
	for i := 0; i < len(errs); i++ {
		temp = errors.Join(temp, errs[i])
	}
	return temp.Error()
}
