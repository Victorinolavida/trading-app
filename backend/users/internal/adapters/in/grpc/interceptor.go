package grpc

import (
	"context"
	"log/slog"

	"github.com/victorinolavida/auth/internal/adapters/in/grpc/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorMappingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		mapped := mapError(err)
		st, ok := status.FromError(mapped)
		code := codes.Unknown
		if ok {
			code = st.Code()
		}

		base := []any{
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
		}
		if reqID := requestid.FromContext(ctx); reqID != "" {
			base = append(base, slog.String("request_id", reqID))
		}

		switch code {
		case codes.Internal:
			base = append(base, slog.Any("error", err))
			slog.Error("grpc request failed", base...)
		case codes.InvalidArgument, codes.AlreadyExists:
			base = append(base, slog.Any("error", err))
			slog.Warn("grpc request rejected", base...)
		case codes.Unauthenticated, codes.PermissionDenied:
			slog.Warn("grpc auth failure", base...)
		default:
			slog.Debug("grpc request error", base...)
		}

		return nil, mapped
	}

	return resp, nil
}
