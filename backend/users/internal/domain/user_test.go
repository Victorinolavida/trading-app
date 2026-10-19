package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserEntity(t *testing.T) {
	testCases := []struct {
		Email     string
		FirstName string
		LastName  string
		Valid     bool
		Reason    string
	}{
		{
			Email:     "",
			FirstName: "Jhon",
			LastName:  "Rod",
			Valid:     false,
			Reason:    "email is empty",
		},
		{
			Email:     "mymail",
			FirstName: "Jhon",
			LastName:  "Rod",
			Valid:     false,
			Reason:    "email is invalid",
		},
		{
			Email:     "jhon.rod@test.com",
			FirstName: "s",
			LastName:  "Rod",
			Valid:     false,
			Reason:    "first name is too short",
		},

		{
			Email:     "jhon.rod@test.com",
			FirstName: "Jhon",
			LastName:  "Ro",
			Valid:     false,
			Reason:    "last name is to short",
		},

		{
			Email:     "jhon.rod@test.com",
			FirstName: "Jhon",
			LastName:  "Rod",
			Valid:     true,
			Reason:    "valid data for a new user",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Reason, func(t *testing.T) {
			_, err := NewUser(tt.Email, tt.FirstName, tt.LastName)
			isValid := err == nil
			t.Logf("error %s", err)

			assert.Equal(
				t,
				tt.Valid,
				isValid,
				"reason: %s, error: %v",
				tt.Reason,
				err,
			)
		})
	}
}
