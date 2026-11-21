package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/auth/internal/application/ports"
	authpb "github.com/victorinolavida/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type resetStub struct{ requestErr, confirmErr error }

func (s resetStub) Request(context.Context, string) error         { return s.requestErr }
func (s resetStub) Confirm(context.Context, string, string) error { return s.confirmErr }

func TestPasswordResetUnavailable(t *testing.T) {
	h := NewAuthGRPCHandler(nil, nil, nil, nil)
	_, err := h.RequestPasswordReset(context.Background(), &authpb.RequestPasswordResetRequest{})
	require.Equal(t, codes.Unimplemented, status.Code(mapError(err)))
	_, err = h.ConfirmPasswordReset(context.Background(), &authpb.ConfirmPasswordResetRequest{})
	require.Equal(t, codes.Unimplemented, status.Code(mapError(err)))
}

func TestPasswordResetRequestDoesNotRevealDeliveryErrors(t *testing.T) {
	h := NewAuthGRPCHandler(nil, nil, nil, resetStub{requestErr: errors.New("delivery failure")})
	response, err := h.RequestPasswordReset(context.Background(), &authpb.RequestPasswordResetRequest{Email: "user@example.com"})
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestPasswordResetInvalidToken(t *testing.T) {
	h := NewAuthGRPCHandler(nil, nil, nil, resetStub{confirmErr: ports.ErrInvalidResetToken})
	_, err := h.ConfirmPasswordReset(context.Background(), &authpb.ConfirmPasswordResetRequest{})
	require.Equal(t, codes.InvalidArgument, status.Code(mapError(err)))
}
