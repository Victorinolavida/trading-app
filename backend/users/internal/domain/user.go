package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	id        uuid.UUID
	email     Email
	firstName Name
	lastName  Name
	createdAt time.Time
	updatedAt time.Time
	password  Password
}

func (u User) ID() uuid.UUID        { return u.id }
func (u User) Email() Email         { return u.email }
func (u User) FirstName() Name      { return u.firstName }
func (u User) LastName() Name       { return u.lastName }
func (u User) CreatedAt() time.Time { return u.createdAt }
func (u User) UpdatedAt() time.Time { return u.updatedAt }
func (u User) Hash() []byte         { return u.password.Hash() }
func (u *User) touch()              { u.updatedAt = time.Now() }

func NewUser(email, firstName, lastName, plainPassword string) (User, error) {
	var user User
	id, err := uuid.NewV7()
	if err != nil {
		return user, err
	}
	e, err := NewEmail(email)
	if err != nil {
		return user, err
	}

	firstNameVo, err := NewName(firstName)
	if err != nil {
		return user, fmt.Errorf("first name %w", err)
	}

	lastNameVo, err := NewName(lastName)
	if err != nil {
		return user, fmt.Errorf("last name %w", err)
	}

	password, err := NewPassword(plainPassword)
	if err != nil {
		return user, err
	}
	now := time.Now()
	return User{
		id:        id,
		createdAt: now,
		updatedAt: now,
		email:     e,
		firstName: firstNameVo,
		lastName:  lastNameVo,
		password:  password,
	}, nil
}

func (u *User) UpdateName(firstName, lastName string) error {
	firstNameVo, err := NewName(firstName)
	if err != nil {
		return fmt.Errorf("first name %w", err)
	}

	lastNameVo, err := NewName(lastName)
	if err != nil {
		return fmt.Errorf("last name %w", err)
	}

	u.lastName = lastNameVo
	u.firstName = firstNameVo
	u.touch()
	return nil
}

func Reconstruct(id uuid.UUID, email, firstName, lastName string, passwordHash []byte, createdAt, updatedAt time.Time) (User, error) {
	firstNameVo, err := NewName(firstName)
	if err != nil {
		return User{}, fmt.Errorf("parsing data from db: %w", err)
	}

	lastNameVo, err := NewName(lastName)
	if err != nil {
		return User{}, fmt.Errorf("parsing data from db: %w", err)
	}

	emailVo, err := NewEmail(email)
	if err != nil {
		return User{}, fmt.Errorf("parsing data from db: %w", err)
	}

	return User{
		id:        id,
		email:     emailVo,
		firstName: firstNameVo,
		lastName:  lastNameVo,
		password:  ReconstructPassword(passwordHash),
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}
