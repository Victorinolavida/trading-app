package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"users/config"
	"users/internal/application/ports"
	"users/internal/domainerrors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time // injectable for tests
}

var (
	_ ports.TokenIssuer   = (*TokenService)(nil)
	_ ports.TokenVerifier = (*TokenService)(nil)
)

type claims struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

func NewService(cfg config.TokenConfig) (*TokenService, error) {
	switch {
	case len(cfg.Secret) < 32:
		return nil, errors.New("jwt: secret must be at least 32 bytes")
	case cfg.Issuer == "" || cfg.Audience == "":
		return nil, errors.New("jwt: issuer and audience are required")
	case cfg.TTL <= 0:
		return nil, errors.New("jwt: ttl must be positive")
	}
	return &TokenService{
		secret:   []byte(cfg.Secret),
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		ttl:      cfg.TTL,
		now:      time.Now,
	}, nil
}

func (serv *TokenService) Issue(_ context.Context, p ports.Claims) (string, error) {
	jti := uuid.New()
	now := time.Now()
	exp := now.Add(serv.ttl)

	c := claims{
		UserID:    p.UserID,
		LastName:  p.LastName,
		FirstName: p.FirstName,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			Subject:   p.UserID,
			Issuer:    serv.issuer,
			Audience:  jwt.ClaimStrings{serv.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(serv.secret)
	if err != nil {
		return "", fmt.Errorf("jwt: sign: %w", err)
	}
	return signed, nil
}

func (serv *TokenService) Verify(ctx context.Context, raw string) (ports.Claims, error) {
	c := &claims{}
	_, err := jwt.ParseWithClaims(raw, c,
		func(*jwt.Token) (any, error) { return serv.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), // pin the algorithm
		jwt.WithIssuer(serv.issuer),
		jwt.WithAudience(serv.audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(5*time.Second),
		jwt.WithTimeFunc(serv.now),
	)
	if err != nil || c.Subject == "" {
		return ports.Claims{}, domainerrors.ErrUnauthenticated
	}

	return ports.Claims{UserID: c.Subject, FirstName: c.FirstName, LastName: c.LastName}, nil
}
