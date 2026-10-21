package response

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapGRPCError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "invalid argument maps to bad request",
			err:        status.Error(codes.InvalidArgument, "missing required fields"),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "missing required fields",
		},
		{
			name:       "unauthenticated maps to unauthorized",
			err:        status.Error(codes.Unauthenticated, "invalid credentials"),
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "invalid credentials",
		},
		{
			name:       "already exists maps to conflict",
			err:        status.Error(codes.AlreadyExists, "email already registered"),
			wantStatus: http.StatusConflict,
			wantMsg:    "email already registered",
		},
		{
			name:       "not found maps to not found",
			err:        status.Error(codes.NotFound, "user not found"),
			wantStatus: http.StatusNotFound,
			wantMsg:    "user not found",
		},
		{
			name:       "permission denied maps to forbidden",
			err:        status.Error(codes.PermissionDenied, "forbidden"),
			wantStatus: http.StatusForbidden,
			wantMsg:    "forbidden",
		},
		{
			name:       "unavailable maps to service unavailable",
			err:        status.Error(codes.Unavailable, "connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantMsg:    "upstream service unavailable",
		},
		{
			name:       "internal error does not leak internals",
			err:        status.Error(codes.Internal, "connection refused on 10.0.0.1:5432"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "something went wrong",
		},
		{
			name:       "context deadline maps to gateway timeout",
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusGatewayTimeout,
			wantMsg:    "upstream service timeout",
		},
		{
			name:       "non grpc error maps to bad gateway",
			err:        errors.New("boom"),
			wantStatus: http.StatusBadGateway,
			wantMsg:    "upstream service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotMsg := MapGRPCError(tt.err)
			if gotStatus != tt.wantStatus {
				t.Errorf("MapGRPCError() status = %d, want %d", gotStatus, tt.wantStatus)
			}
			if gotMsg != tt.wantMsg {
				t.Errorf("MapGRPCError() message = %q, want %q", gotMsg, tt.wantMsg)
			}
		})
	}
}
