package metadata

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type ctxKey string

const (
	clientIPKey  ctxKey = "client_ip"
	userAgentKey ctxKey = "user_agent"
	userIDKey    ctxKey = "user_id"
)

func ClientIPFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(clientIPKey).(string); ok {
		return v
	}
	return ""
}

func UserAgentFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userAgentKey).(string); ok {
		return v
	}
	return ""
}

func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

func inject(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	if vals := md.Get("x-client-ip"); len(vals) > 0 && vals[0] != "" {
		ctx = context.WithValue(ctx, clientIPKey, vals[0])
	}
	if vals := md.Get("user-agent"); len(vals) > 0 && vals[0] != "" {
		ctx = context.WithValue(ctx, userAgentKey, vals[0])
	}
	if vals := md.Get("x-user-id"); len(vals) > 0 && vals[0] != "" {
		ctx = context.WithValue(ctx, userIDKey, vals[0])
	}
	return ctx
}
