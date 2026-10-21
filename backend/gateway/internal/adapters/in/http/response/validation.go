package response

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// NewValidator returns a validator that reports the invalid fields with the
// name the client knows (the json tag) instead of the go field name.
func NewValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(jsonTagName)

	return validate
}

// jsonTagName returns the json name of a struct field.
func jsonTagName(field reflect.StructField) string {
	name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

// RenderValidationError writes a client safe representation of a validation
// error. It never exposes the internals of the validation library.
func RenderValidationError(w http.ResponseWriter, err error) {
	RenderBadRequest(w, ValidationMessage(err))
}

// ValidationMessage builds a human readable message out of the validation
// errors of a request.
func ValidationMessage(err error) string {
	var invalid validator.ValidationErrors
	if !errors.As(err, &invalid) {
		return "invalid request body"
	}

	messages := make([]string, 0, len(invalid))
	for _, fieldErr := range invalid {
		messages = append(messages, fmt.Sprintf("%q %s", fieldErr.Field(), ruleMessage(fieldErr.Tag())))
	}

	return strings.Join(messages, ", ")
}

// ruleMessage translates a validation rule into a message for the client.
func ruleMessage(tag string) string {
	switch tag {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	default:
		return "is invalid"
	}
}
