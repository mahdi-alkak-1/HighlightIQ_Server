package auth

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func (r LoginRequest) Validate() error {
	clean := r
	clean.Email = strings.TrimSpace(clean.Email)

	if err := validate.Struct(clean); err != nil {
		errs := ValidationError{}

		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			errs["general"] = "invalid request"
			return errs
		}

		for _, fe := range ve {
			key := jsonFieldNameLogin(clean, fe.Field())
			errs[key] = messageForFieldError(key, fe)
		}

		if len(errs) > 0 {
			return errs
		}
	}

	return nil
}

func jsonFieldNameLogin(req LoginRequest, structFieldName string) string {
	t := reflect.TypeOf(req)
	f, ok := t.FieldByName(structFieldName)
	if !ok {
		return strings.ToLower(structFieldName)
	}

	tag := f.Tag.Get("json")
	if tag == "" {
		return strings.ToLower(structFieldName)
	}

	if i := strings.Index(tag, ","); i >= 0 {
		tag = tag[:i]
	}
	if tag == "" || tag == "-" {
		return strings.ToLower(structFieldName)
	}
	return tag
}
