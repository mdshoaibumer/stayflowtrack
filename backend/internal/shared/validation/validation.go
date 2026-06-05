package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Validate(s any) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrors []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   toSnakeCase(e.Field()),
			Message: msgForTag(e),
		})
	}
	return fieldErrors
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "e164":
		return "must be a valid phone number"
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
