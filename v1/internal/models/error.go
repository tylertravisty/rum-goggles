package models

import "fmt"

const (
	pkgName = "models"

	ErrAccountInvalidUsername ValidatorError = "invalid account username"
	ErrAccountInvalidID       ValidatorError = "invalid account id"
)

func pkgErr(prefix string, err error) error {
	pkgErr := pkgName
	if prefix != "" {
		pkgErr = fmt.Sprintf("%s: %s", pkgErr, prefix)
	}

	return fmt.Errorf("%s: %v", pkgErr, err)
}

type ValidatorError string

func (e ValidatorError) Error() string {
	return string(e)
}
