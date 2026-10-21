package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/victorinolavida/auth/config"
	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

const testSecret = "test-secret-0123456789-abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJ"

func validConfig() config.TokenConfig {
	return config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test-issuer",
		Audience: "test-audience",
		TTL:      15 * time.Minute,
	}
}

func newTestService(t *testing.T) *TokenService {
	t.Helper()
	serv, err := NewService(validConfig())
	assert.NoError(t, err)
	return serv
}

func TestNewService(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*config.TokenConfig)
		wantErr string
	}{
		{name: "valid config"},
		{
			name:    "secret too short",
			mutate:  func(c *config.TokenConfig) { c.Secret = "too-short" },
			wantErr: "secret",
		},
		{
			name:    "missing issuer",
			mutate:  func(c *config.TokenConfig) { c.Issuer = "" },
			wantErr: "issuer",
		},
		{
			name:    "missing audience",
			mutate:  func(c *config.TokenConfig) { c.Audience = "" },
			wantErr: "audience",
		},
		{
			name:    "zero ttl",
			mutate:  func(c *config.TokenConfig) { c.TTL = 0 },
			wantErr: "ttl",
		},
		{
			name:    "negative ttl",
			mutate:  func(c *config.TokenConfig) { c.TTL = -time.Minute },
			wantErr: "ttl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			if tt.mutate != nil {
				tt.mutate(&cfg)
			}

			serv, err := NewService(cfg)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Nil(t, serv)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, serv)
		})
	}
}

func TestIssueAndVerify_HappyPath(t *testing.T) {
	serv := newTestService(t)
	want := ports.Claims{
		UserID:    "user-123",
		FirstName: "John",
		LastName:  "Doe",
	}

	raw, err := serv.Issue(context.Background(), want)
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)

	got, err := serv.Verify(context.Background(), raw)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestIssue_RegisteredClaims(t *testing.T) {
	serv := newTestService(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	serv.now = func() time.Time { return now }

	raw, err := serv.Issue(context.Background(), ports.Claims{UserID: "user-123"})
	assert.NoError(t, err)

	parsed := &claims{}
	_, err = jwt.ParseWithClaims(raw, parsed, func(*jwt.Token) (any, error) {
		return serv.secret, nil
	}, jwt.WithTimeFunc(serv.now))
	assert.NoError(t, err)

	assert.Equal(t, "test-issuer", parsed.Issuer)
	assert.Equal(t, jwt.ClaimStrings{"test-audience"}, parsed.Audience)
	assert.Equal(t, "user-123", parsed.Subject)
	assert.Equal(t, "user-123", parsed.UserID)
	assert.NotEmpty(t, parsed.ID)
	assert.Equal(t, now.Unix(), parsed.IssuedAt.Unix())
	assert.Equal(t, now.Unix(), parsed.NotBefore.Unix())
	assert.Equal(t, now.Add(15*time.Minute).Unix(), parsed.ExpiresAt.Unix())
}

func TestVerify_Expired(t *testing.T) {
	serv := newTestService(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	serv.now = func() time.Time { return now }

	raw, err := serv.Issue(context.Background(), ports.Claims{UserID: "user-123"})
	assert.NoError(t, err)

	serv.now = func() time.Time { return now.Add(15*time.Minute + 6*time.Second) }

	_, err = serv.Verify(context.Background(), raw)
	assert.ErrorIs(t, err, domainerrors.ErrUnauthenticated)
}

func TestVerify_ExpiredWithinLeeway(t *testing.T) {
	serv := newTestService(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	serv.now = func() time.Time { return now }

	raw, err := serv.Issue(context.Background(), ports.Claims{UserID: "user-123"})
	assert.NoError(t, err)

	serv.now = func() time.Time { return now.Add(15*time.Minute + 3*time.Second) }

	_, err = serv.Verify(context.Background(), raw)
	assert.NoError(t, err)
}

func TestVerify_RejectsInvalidTokens(t *testing.T) {
	serv := newTestService(t)
	valid := claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    serv.issuer,
			Audience:  jwt.ClaimStrings{serv.audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	noSubject := valid
	noSubject.Subject = ""

	noneToken, err := jwt.NewWithClaims(jwt.SigningMethodNone, valid).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	hs512Token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, valid).
		SignedString(serv.secret)
	assert.NoError(t, err)

	noSubjectToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, noSubject).
		SignedString(serv.secret)
	assert.NoError(t, err)

	wrongIssuer := valid
	wrongIssuer.Issuer = "someone-else"
	wrongIssuerToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, wrongIssuer).
		SignedString(serv.secret)
	assert.NoError(t, err)

	wrongAudience := valid
	wrongAudience.Audience = jwt.ClaimStrings{"someone-else"}
	wrongAudienceToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, wrongAudience).
		SignedString(serv.secret)
	assert.NoError(t, err)

	noExpiry := valid
	noExpiry.ExpiresAt = nil
	noExpiryToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, noExpiry).
		SignedString(serv.secret)
	assert.NoError(t, err)

	otherServ, err := NewService(config.TokenConfig{
		Secret:   "another-test-secret-0123456789-abcdefghijklmnopqrstuvwxyz",
		Issuer:   serv.issuer,
		Audience: serv.audience,
		TTL:      time.Minute,
	})
	assert.NoError(t, err)
	wrongSecretToken, err := otherServ.Issue(context.Background(), ports.Claims{UserID: "user-123"})
	assert.NoError(t, err)

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty", token: ""},
		{name: "garbage", token: "not-a-token"},
		{name: "alg none", token: noneToken},
		{name: "hs512 instead of hs256", token: hs512Token},
		{name: "missing subject", token: noSubjectToken},
		{name: "wrong issuer", token: wrongIssuerToken},
		{name: "wrong audience", token: wrongAudienceToken},
		{name: "missing expiry", token: noExpiryToken},
		{name: "signed with wrong secret", token: wrongSecretToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := serv.Verify(context.Background(), tt.token)
			assert.ErrorIs(t, err, domainerrors.ErrUnauthenticated)
		})
	}
}
