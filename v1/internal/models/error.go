package models

import "fmt"

const (
	pkgName = "models"

	ErrAccountInvalidUsername ValidatorError = "invalid account username"
	ErrAccountInvalidID       ValidatorError = "invalid account id"

	ErrChannelInvalidAccountID ValidatorError = "invalid channel account id"
	ErrChannelInvalidApiKey    ValidatorError = "invalid channel API key"
	ErrChannelInvalidCID       ValidatorError = "invalid channel CID"
	ErrChannelInvalidName      ValidatorError = "invalid channel name"
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
