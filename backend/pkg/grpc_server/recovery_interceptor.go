package grpcserver

import (
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func recoveryInterceptor() grpc.UnaryServerInterceptor {
	grpcPanicRecoveryHandler := func(p any) (err error) {
		slog.Error("recovered from panic", slog.Any("panic", p))
		return status.Error(codes.Internal, "internal error")
	}
	return recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler))
}
