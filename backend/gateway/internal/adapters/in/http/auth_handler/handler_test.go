package authhandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	pkgconfig "github.com/victorinolavida/pkg/config"
)

type stubAuthService struct {
	loginResult    dtos.AuthUserResponse
	loginErr       error
	registerResult dtos.AuthUserResponse
	registerErr    error
	whoamiResult   dtos.WhoamiUserResponse
	whoamiErr      error
}

func (s *stubAuthService) Login(context.Context, dtos.LoginRequest) (dtos.AuthUserResponse, error) {
	return s.loginResult, s.loginErr
}

func (s *stubAuthService) Register(context.Context, dtos.RegisterRequest) (dtos.AuthUserResponse, error) {
	return s.registerResult, s.registerErr
}

func (s *stubAuthService) Whoami(context.Context, dtos.WhoamiUserRequest) (dtos.WhoamiUserResponse, error) {
	return s.whoamiResult, s.whoamiErr
}

func validTestConfig() *config.Config {
	return &config.Config{
		Token: pkgconfig.TokenConfig{
			Secret:   "test-secret-0123456789-abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJ",
			Issuer:   "test-issuer",
			Audience: "test-audience",
			TTL:      15 * 60 * 1e9,
		},
	}
}

func TestNewAuthHandler_MissingValidator(t *testing.T) {
	handler, err := NewAuthHandler(validTestConfig(), &stubAuthService{}, nil)

	assert.ErrorIs(t, err, ErrMissingValidator)
	assert.Nil(t, handler)
}

func TestNewAuthHandler(t *testing.T) {
	handler, err := NewAuthHandler(validTestConfig(), &stubAuthService{}, response.NewValidator())

	require.NoError(t, err)
	assert.NotNil(t, handler)
}

func TestAuthHandler_Login(t *testing.T) {
	service := &stubAuthService{
		loginResult: dtos.AuthUserResponse{
			UserID:    "user-1",
			FirstName: "Jane",
			LastName:  "Doe",
			Token:     "a-token",
		},
	}
	handler, err := NewAuthHandler(validTestConfig(), service, response.NewValidator())
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"jane@example.com","password":"secret"}`))

	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body AuthUserResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	assert.Equal(t, AuthUserResponse{
		UserID:    "user-1",
		FirstName: "Jane",
		LastName:  "Doe",
		Token:     "a-token",
	}, body)
}

func TestAuthHandler_LoginValidationError(t *testing.T) {
	handler, err := NewAuthHandler(validTestConfig(), &stubAuthService{}, response.NewValidator())
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"not-an-email","password":""}`))

	handler.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "must be a valid email")
	assert.Contains(t, rr.Body.String(), "is required")
}

func TestAuthHandler_LoginUpstreamError(t *testing.T) {
	service := &stubAuthService{loginErr: errors.New("boom")}
	handler, err := NewAuthHandler(validTestConfig(), service, response.NewValidator())
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"jane@example.com","password":"secret"}`))

	handler.Login(rr, req)

	// a non gRPC error is a broken upstream, not a bug of the gateway
	assert.Equal(t, http.StatusBadGateway, rr.Code)
}

func TestAuthHandler_Routes(t *testing.T) {
	handler, err := NewAuthHandler(validTestConfig(), &stubAuthService{}, response.NewValidator())
	require.NoError(t, err)

	assert.Equal(t, "/auth", handler.BasePath())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"jane@example.com","password":"secret"}`))

	handler.Routes().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

var _ ports.AuthService = (*stubAuthService)(nil)
