// Package api is the public boundary of the users bounded context.
//
// The rest of the application talks to the users module through this package
// instead of importing its internal packages. Keeping the boundary explicit is
// what allows the module to run in-process today and be extracted into a
// separate service later without touching the callers.
package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/victorinolavida/auth/internal/adapters/out/postgres"
	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/application/usecases/login"
	"github.com/victorinolavida/auth/internal/application/usecases/register"
	"github.com/victorinolavida/auth/internal/application/usecases/whoami"
	pkgconfig "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/jwt"
)

// User is the user representation exposed across module boundaries.
type User struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

// AuthResult is returned by the login and register use cases.
type AuthResult struct {
	User  User
	Token string
}

// Module is the in-process users module.
type Module struct {
	repo     ports.UserRepository
	login    *login.Usecase
	register *register.Usecase
	whoami   *whoami.Usecase
}

// New wires the users module on top of the given database pool.
func New(pool *pgxpool.Pool, tokenCfg pkgconfig.TokenConfig) (*Module, error) {
	issuer, err := jwt.NewService(tokenCfg)
	if err != nil {
		return nil, err
	}

	repo := postgres.NewUserRepository(pool)

	return &Module{
		repo:     repo,
		login:    login.New(repo, issuer),
		register: register.New(repo, issuer),
		whoami:   whoami.New(repo),
	}, nil
}

// Login authenticates a user and returns a signed token.
func (m *Module) Login(ctx context.Context, email, password string) (AuthResult, error) {
	res, err := m.login.Execute(ctx, login.Command{Email: email, Password: password})
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		User: User{
			ID:        res.User.ID,
			Email:     res.User.Email,
			FirstName: res.User.FirstName,
			LastName:  res.User.LastName,
		},
		Token: res.Token,
	}, nil
}

// Register creates a user and returns a signed token.
func (m *Module) Register(ctx context.Context, email, firstName, lastName, password string) (AuthResult, error) {
	res, err := m.register.Execute(ctx, register.Command{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	})
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		User: User{
			ID:        res.User.ID,
			Email:     res.User.Email,
			FirstName: res.User.FirstName,
			LastName:  res.User.LastName,
		},
		Token: res.Token,
	}, nil
}

// Whoami resolves the public profile of a user id.
func (m *Module) Whoami(ctx context.Context, userID string) (User, error) {
	res, err := m.whoami.Execute(ctx, whoami.Command{UserID: userID})
	if err != nil {
		return User{}, err
	}
	return User{
		ID:        res.UserID,
		Email:     res.Email,
		FirstName: res.FirstName,
		LastName:  res.LastName,
	}, nil
}

// GetByID resolves a user by its internal id.
func (m *Module) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	user, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:        user.ID().String(),
		Email:     user.Email().String(),
		FirstName: user.FirstName().String(),
		LastName:  user.LastName().String(),
	}, nil
}
