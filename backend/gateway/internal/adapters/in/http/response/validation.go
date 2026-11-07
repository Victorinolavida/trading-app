package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// jsonTagName returns the json name of a struct field, falling back to the
// Go field name when the field is not exposed in JSON.
func jsonTagName(field reflect.StructField) string {
	name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	if name == "-" || name == "" {
		return field.Name
	}
	return name
}

// RenderValidationError writes a client safe representation of a validation
// error. It never exposes the internals of the validation library.
func RenderValidationError(w http.ResponseWriter, err error) {
	RenderBadRequest(w, ValidationMessage(err))
}

// DecodeErrorMessage builds a human readable, client safe message out of a
// JSON body decoding error. It never exposes Go struct or type internals,
// e.g. a quantity sent as a string becomes "invalid quantity field".
func DecodeErrorMessage(err error) string {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := typeErr.Field
		if i := strings.LastIndex(field, "."); i >= 0 {
			field = field[i+1:]
		}
		if field == "" {
			return "invalid body"
		}
		return fmt.Sprintf("invalid %s field", field)
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return "request body too large"
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "invalid body"
	}

	// json.SyntaxError and unknown-field errors have no dedicated type.
	if msg := err.Error(); strings.HasPrefix(msg, "json: unknown field ") {
		return strings.TrimPrefix(msg, "json: ")
	}

	return "invalid body"
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
