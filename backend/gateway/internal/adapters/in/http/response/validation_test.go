package response

import (
	"net/http/httptest"
	"testing"
)

type validationPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func TestValidationMessage(t *testing.T) {
	validate := NewValidator()

	tests := []struct {
		name    string
		payload validationPayload
		want    string
	}{
		{
			name:    "missing every field",
			payload: validationPayload{},
			want:    `"email" is required, "password" is required`,
		},
		{
			name:    "invalid email",
			payload: validationPayload{Email: "not-an-email", Password: "secret"},
			want:    `"email" must be a valid email`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.payload)
			if err == nil {
				t.Fatal("expected a validation error, got nil")
			}

			got := ValidationMessage(err)
			if got != tt.want {
				t.Errorf("ValidationMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderValidationError(t *testing.T) {
	validate := NewValidator()

	err := validate.Struct(validationPayload{})
	if err == nil {
		t.Fatal("expected a validation error, got nil")
	}

	rec := httptest.NewRecorder()
	RenderValidationError(rec, err)

	if rec.Code != 400 {
		t.Errorf("status = %d, want %d", rec.Code, 400)
	}
}
