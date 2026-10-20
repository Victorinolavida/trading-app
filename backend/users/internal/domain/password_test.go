package domain

import (
	"strings"
	"sync"
	"testing"
	"users/internal/domainerrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const fixturePlain = "Sup3rSecret"

var (
	fixtureOnce sync.Once
	fixturePass Password
	fixtureErr  error
)

func fixturePassword(t *testing.T) Password {
	t.Helper()

	fixtureOnce.Do(func() {
		fixturePass, fixtureErr = NewPassword(fixturePlain)
	})

	require.NoError(t, fixtureErr)
	return fixturePass
}

func TestNewPasswordValidation(t *testing.T) {
	testCases := []struct {
		Plain  string
		Valid  bool
		Err    error
		Reason string
	}{
		{Plain: "", Valid: false, Err: domainerrors.ErrPasswordTooShort, Reason: "empty password"},
		{Plain: "Ab1", Valid: false, Err: domainerrors.ErrPasswordTooShort, Reason: "below minimum length"},
		{Plain: "Ab1defg", Valid: false, Err: domainerrors.ErrPasswordTooShort, Reason: "one byte below minimum length"},
		{Plain: "Ab1defgh", Valid: true, Reason: "exactly minimum length"},
		{Plain: strings.Repeat("a", 23), Valid: true, Reason: "one byte below maximum length"},
		{Plain: strings.Repeat("a", 24), Valid: false, Err: domainerrors.ErrPasswordTooLong, Reason: "exactly maximum length"},
		{Plain: strings.Repeat("a", 64), Valid: false, Err: domainerrors.ErrPasswordTooLong, Reason: "above maximum length"},
	}

	for _, tt := range testCases {
		t.Run(tt.Reason, func(t *testing.T) {
			_, err := NewPassword(tt.Plain)

			if tt.Valid {
				require.NoError(t, err, "Password: %q", tt.Plain)
				return
			}

			require.Error(t, err, "Password: %q", tt.Plain)
			assert.ErrorIs(t, err, tt.Err, "Password: %q", tt.Plain)
		})
	}
}

func TestPasswordHashIsNotPlaintextAndUsesDomainCost(t *testing.T) {
	p := fixturePassword(t)

	assert.NotEqual(t, fixturePlain, string(p.Hash()), "hash must not be the plaintext password")
	assert.NotEmpty(t, p.Hash())

	hashCost, err := bcrypt.Cost(p.Hash())
	require.NoError(t, err, "hash must be a valid bcrypt hash")
	assert.Equal(t, cost, hashCost, "hash must use the domain configured cost")
}

func TestPasswordHashesAreSalted(t *testing.T) {
	first, err := NewPassword(fixturePlain)
	require.NoError(t, err)

	second, err := NewPassword(fixturePlain)
	require.NoError(t, err)

	assert.NotEqual(t, string(first.Hash()), string(second.Hash()), "same plaintext must produce different hashes (random salt)")
}

func TestPasswordVerify(t *testing.T) {
	p := fixturePassword(t)

	testCases := []struct {
		Plain  string
		Match  bool
		Reason string
	}{
		{Plain: fixturePlain, Match: true, Reason: "same password"},
		{Plain: "Sup3rSecret2", Match: false, Reason: "different password"},
		{Plain: "sup3rsecret", Match: false, Reason: "same password with different case"},
		{Plain: "", Match: false, Reason: "empty password"},
	}

	for _, tt := range testCases {
		t.Run(tt.Reason, func(t *testing.T) {
			assert.Equal(t, tt.Match, p.Verify(tt.Plain))
		})
	}
}

func TestVerifyPasswordAgainstStoredHash(t *testing.T) {
	p := fixturePassword(t)
	storedHash := string(p.Hash())

	assert.True(t, VerifyPassword(fixturePlain, storedHash))
	assert.False(t, VerifyPassword("Sup3rSecret2", storedHash))
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	testCases := []struct {
		Hash   string
		Reason string
	}{
		{Hash: "", Reason: "empty hash"},
		{Hash: "not-a-bcrypt-hash", Reason: "malformed hash"},
		{Hash: "$2a$14$invalid", Reason: "truncated bcrypt hash"},
	}

	for _, tt := range testCases {
		t.Run(tt.Reason, func(t *testing.T) {
			assert.False(t, VerifyPassword(fixturePlain, tt.Hash))
		})
	}
}
