package response

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDecodeErrorMessage(t *testing.T) {
	type payload struct {
		Quantity int64 `json:"quantity"`
	}

	var p payload
	typeErr := json.Unmarshal([]byte(`{"quantity":"100"}`), &p)

	unknownDecoder := json.NewDecoder(strings.NewReader(`{"quantit":1}`))
	unknownDecoder.DisallowUnknownFields()
	var u payload
	unknownErr := unknownDecoder.Decode(&u)

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"wrong quantity type", typeErr, "invalid quantity field"},
		{"unknown field", unknownErr, `unknown field "quantit"`},
		{"syntax error", json.Unmarshal([]byte(`{"quantity":`), &p), "invalid body"},
		{"generic error", errors.New("boom"), "invalid body"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecodeErrorMessage(tt.err); got != tt.want {
				t.Errorf("DecodeErrorMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
