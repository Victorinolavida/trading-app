package metadata

import (
	"context"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	ctx = inject(ctx)
	return handler(ctx, req)
}
