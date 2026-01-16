package recordings

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type UpdateTitleRequest struct {
	Title string `json:"title" validate:"required,max=120"`
}

type ValidationError map[string]string

func (e ValidationError) Error() string {
	return "validation error"
}

func (r UpdateTitleRequest) Validate() error {
	clean := r
	clean.Title = strings.TrimSpace(clean.Title)

	if err := validate.Struct(clean); err != nil {
		errs := ValidationError{}

		ve, ok := err.(validator.ValidationErrors)
		if !ok {
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
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	default:
		return field + " is invalid"
	}
}

func jsonFieldName(req UpdateTitleRequest, structFieldName string) string {
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
