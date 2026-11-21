package grpcserver

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRequestErrorInterceptorLogsRequestID(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "request-123"))
	want := status.Error(codes.Internal, "failure")
	_, err := requestErrorInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/service/method"}, func(context.Context, any) (any, error) { return nil, want })
	if err != want {
		t.Fatalf("error = %v, want %v", err, want)
	}
	if !strings.Contains(logs.String(), `"request_id":"request-123"`) {
		t.Fatalf("missing request ID in log: %s", logs.String())
	}
}
