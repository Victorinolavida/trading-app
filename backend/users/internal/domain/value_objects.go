package domain

import (
	"regexp"
	"users/internal/domainerrors"
)

var emailRegExpression = regexp.MustCompile(
	`^[\w-]+(?:\.[\w-]+)*@([\w-]+\.)+[a-zA-Z]{2,4}$`,
)

const (
	minNameLength = 3
	maxNameLength = 50
)

type Email struct {
	value string
}

func (e Email) validate() error {
	if !emailRegExpression.MatchString(e.value) {
		return domainerrors.ErrInvalidEmail
	}
	return nil
}

func (e Email) String() string { return e.value }

func NewEmail(value string) (Email, error) {
	email := Email{value: value}
	if err := email.validate(); err != nil {
		return Email{}, err
	}
	return email, nil
}

type Name struct {
	value string
}

func (n Name) String() string { return n.value }

func NewName(value string) (Name, error) {
	n := Name{value}
	if len(n.value) < minNameLength {
		return Name{}, domainerrors.ErrNameTooShort
	}

	if len(n.value) >= maxNameLength {
		return Name{}, domainerrors.ErrNameTooLong
	}

	return n, nil
}
