package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     Email
	FirstName Name
	LastName  Name
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email, firstName, lastName string) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	e, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	firstNameVo, err := NewName(firstName)
	if err != nil {
		return nil, fmt.Errorf("first name %w", err)
	}

	LastNameVo, err := NewName(lastName)
	if err != nil {
		return nil, fmt.Errorf("last name %w", err)
	}

	return &User{
		ID:        id,
		CreatedAt: time.Now(),
		Email:     e,
		FirstName: firstNameVo,
		LastName:  LastNameVo,
	}, nil
}
