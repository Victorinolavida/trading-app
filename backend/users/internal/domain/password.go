package domain

import (
	"github.com/victorinolavida/auth/internal/domainerrors"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 24
	defaultCost       = 14
)

var cost = defaultCost

type Password struct {
	hash []byte
}

func (p Password) Hash() []byte {
	return p.hash
}

func (p Password) Verify(plain string) bool {
	return VerifyPassword(plain, string(p.hash))
}

func NewPassword(plain string) (Password, error) {
	if len(plain) < minPasswordLength {
		return Password{}, domainerrors.ErrPasswordTooShort
	}

	if len(plain) >= maxPasswordLength {
		return Password{}, domainerrors.ErrPasswordTooLong
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return Password{}, err
	}
	return Password{hash: hash}, nil
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ReconstructPassword(hash []byte) Password {
	return Password{hash: append([]byte(nil), hash...)}
}
