package grpcserver

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func requestErrorInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		code := status.Code(err)
		args := []any{slog.String("method", info.FullMethod), slog.String("code", code.String())}
		if id := incomingRequestID(ctx); id != "" {
			args = append(args, slog.String("request_id", id))
		}
		if code == codes.Internal || code == codes.Unknown {
			slog.Error("grpc request failed", args...)
		} else {
			slog.Debug("grpc request failed", args...)
		}
	}
	return resp, err
}
