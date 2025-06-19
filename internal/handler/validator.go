// Package handler provides HTTP request validation utilities for the authentication service.
package handler

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	phoneRegexp = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	nameRegexp  = regexp.MustCompile(`^[가-힣a-zA-Z]+$`)

	// Validate is the global validator instance.
	Validate *validator.Validate
)

func init() {
	Validate = validator.New()
	_ = Validate.RegisterValidation("namekr", NameValidator)
	_ = Validate.RegisterValidation("phonekr", PhoneValidator)
}

// PhoneValidator validates phone numbers in E.164 format.
func PhoneValidator(fl validator.FieldLevel) bool {
	return phoneRegexp.MatchString(fl.Field().String())
}

// NameValidator validates Korean and English names.
func NameValidator(fl validator.FieldLevel) bool {
	return nameRegexp.MatchString(fl.Field().String())
}
