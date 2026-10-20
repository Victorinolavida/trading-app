package grpc

import (
	"context"
	"log/slog"

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

		switch code {
		case codes.Internal:
			slog.Error("grpc request failed",
				slog.String("method", info.FullMethod),
				slog.String("code", code.String()),
				slog.Any("error", err),
			)
		case codes.InvalidArgument, codes.AlreadyExists:
			slog.Warn("grpc request rejected",
				slog.String("method", info.FullMethod),
				slog.String("code", code.String()),
				slog.Any("error", err),
			)
		case codes.Unauthenticated, codes.PermissionDenied:
			slog.Warn("grpc auth failure",
				slog.String("method", info.FullMethod),
				slog.String("code", code.String()),
			)
		default:
			slog.Debug("grpc request error",
				slog.String("method", info.FullMethod),
				slog.String("code", code.String()),
			)
		}

		return nil, mapped
	}

	return resp, nil
}
