package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailVO(t *testing.T) {
	testCases := []struct {
		Email  string
		Valid  bool
		Reason string
	}{
		// Valid email addresses
		{Email: "john@gmail.com", Valid: true, Reason: "Standard email address"},
		{Email: "john.doe@gmail.com", Valid: true, Reason: "Username contains a dot"},
		{Email: "john_doe@gmail.com", Valid: true, Reason: "Username contains an underscore"},
		{Email: "john-doe@gmail.com", Valid: true, Reason: "Username contains a hyphen"},
		{Email: "john123@gmail.com", Valid: true, Reason: "Username contains numbers"},
		{Email: "a@b.co", Valid: true, Reason: "Single-character username and 2-character TLD"},
		{Email: "user@mail.co.uk", Valid: true, Reason: "Domain contains multiple sections"},
		{Email: "user@company.info", Valid: true, Reason: "TLD contains 4 characters"},
		{Email: "user.name+tag@gmail.com", Valid: false, Reason: "Plus sign is not allowed by this regex"},
		{Email: "user%name@gmail.com", Valid: false, Reason: "Percent sign is not allowed by this regex"},

		// Invalid email addresses
		{Email: "@gmail.com", Valid: false, Reason: "Missing username"},
		{Email: "john@gmail", Valid: false, Reason: "Missing domain extension"},
		{Email: "john@gmail.c", Valid: false, Reason: "TLD has fewer than 2 characters"},
		{Email: "john@gmail.museum", Valid: false, Reason: "TLD exceeds the regex's 4-character limit"},
		{Email: "john..doe@gmail.com", Valid: false, Reason: "Consecutive dots in username"},
		{Email: "john@gmail..com", Valid: false, Reason: "Empty domain section caused by consecutive dots"},
		{Email: "john gmail@gmail.com", Valid: false, Reason: "Username contains a space"},
		{Email: "john@gmail.c1", Valid: false, Reason: "TLD contains a number"},
		{Email: "john@.com", Valid: false, Reason: "Missing domain name"},
		{Email: "john@gmail.com extra", Valid: false, Reason: "Contains trailing text after the email"},
		{Email: "", Valid: false, Reason: "Empty input"},
		{Email: "plain-text", Valid: false, Reason: "Missing @ symbol and domain"},
		{Email: "john@", Valid: false, Reason: "Missing domain"},
		{Email: "john@@gmail.com", Valid: false, Reason: "Contains two @ symbols"},
		{Email: ".john@gmail.com", Valid: false, Reason: "Username starts with a dot"},
		{Email: "john@gmail.com.", Valid: false, Reason: "Contains a trailing dot after the TLD"},
	}

	for _, tt := range testCases {
		t.Run(tt.Reason, func(t *testing.T) {
			_, err := NewEmail(tt.Email)
			isValid := err == nil

			assert.Equal(
				t,
				tt.Valid,
				isValid,
				"Email: %q, reason: %s, error: %v",
				tt.Email,
				tt.Reason,
				err,
			)
		})
	}
}

func TestNameVO(t *testing.T) {
	testCases := []struct {
		Name   string
		Valid  bool
		Reason string
	}{
		{
			Name:   "Jhon",
			Valid:  true,
			Reason: "valid name",
		},
		{
			Name:   "",
			Valid:  false,
			Reason: "empty string",
		},

		{
			Name:   "a",
			Valid:  false,
			Reason: "first name to short",
		},

		{
			Name:   "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwx",
			Valid:  false,
			Reason: "first name to long",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Reason, func(t *testing.T) {
			_, err := NewName(tt.Name)
			isValid := err == nil

			assert.Equal(
				t,
				tt.Valid,
				isValid,
				"Name: %q, reason: %s, error: %v",
				tt.Name,
				tt.Reason,
				err,
			)
		})
	}
}
