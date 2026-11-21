package grpcserver

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func recoveryInterceptor() grpc.UnaryServerInterceptor {
	grpcPanicRecoveryHandler := func(ctx context.Context, p any) (err error) {
		args := []any{slog.Any("panic", p), slog.String("stack", string(debug.Stack()))}
		if id := incomingRequestID(ctx); id != "" {
			args = append(args, slog.String("request_id", id))
		}
		slog.Error("recovered from panic", args...)
		return status.Error(codes.Internal, "internal error")
	}
	return recovery.UnaryServerInterceptor(recovery.WithRecoveryHandlerContext(grpcPanicRecoveryHandler))
}

func incomingRequestID(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if ids := md.Get("x-request-id"); len(ids) > 0 {
		return ids[0]
	}
	return ""
}
