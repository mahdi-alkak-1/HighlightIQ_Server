package auth

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func (r RegisterRequest) Validate() error {
	// Validate trimmed values (common UX: " user@email.com " should work)
	clean := r
	clean.Name = strings.TrimSpace(clean.Name)
	clean.Email = strings.TrimSpace(clean.Email)

	if err := validate.Struct(clean); err != nil {
		errs := ValidationError{}

		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			// Unknown validation error type
			errs["general"] = "invalid request"
			return errs
		}

		for _, fe := range ve {
			key := jsonFieldName(clean, fe.Field())
			errs[key] = messageForFieldError(key, fe)
		}

		if len(errs) > 0 {
			return errs
		}
	}

	return nil
}

func messageForFieldError(field string, fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "email":
		return "email is invalid"
	default:
		return field + " is invalid"
	}
}

func jsonFieldName(req RegisterRequest, structFieldName string) string {
	// Map Go struct field name -> json tag (so you get "email" not "Email")
	t := reflect.TypeOf(req)
	f, ok := t.FieldByName(structFieldName)
	if !ok {
		return strings.ToLower(structFieldName)
	}

	tag := f.Tag.Get("json")
	if tag == "" {
		return strings.ToLower(structFieldName)
	}

	// tag might be "email,omitempty"
	if i := strings.Index(tag, ","); i >= 0 {
		tag = tag[:i]
	}
	if tag == "" || tag == "-" {
		return strings.ToLower(structFieldName)
	}
	return tag
}
