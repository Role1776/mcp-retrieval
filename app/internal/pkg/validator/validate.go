package validator

import "github.com/go-playground/validator/v10"

var validate = validator.New(validator.WithRequiredStructEnabled())

func Validate(data any) error {
	return validate.Struct(data)
}
