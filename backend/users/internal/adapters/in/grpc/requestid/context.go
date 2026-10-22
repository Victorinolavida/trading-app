package requestid

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type ctxKey string

const key ctxKey = "request_id"

func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}

func inject(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	vals := md.Get("x-request-id")
	if len(vals) == 0 || vals[0] == "" {
		return ctx
	}
	return context.WithValue(ctx, key, vals[0])
}
