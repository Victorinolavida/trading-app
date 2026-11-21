package grpcclient

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestRequestMetaInterceptor(t *testing.T) {
	for _, tc := range []struct {
		name string
		ctx  context.Context
		want string
	}{
		{"http request", context.WithValue(context.Background(), middleware.RequestIDKey, "http-id"), "http-id"},
		{"service request", metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "service-id")), "service-id"},
		{"no request id", context.Background(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewOutgoingContext(tc.ctx, metadata.Pairs("other", "kept"))
			err := RequestMetaInterceptor(ctx, "/test", nil, nil, nil, func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
				md, _ := metadata.FromOutgoingContext(ctx)
				ids := md.Get("x-request-id")
				if tc.want == "" && len(ids) != 0 || tc.want != "" && (len(ids) != 1 || ids[0] != tc.want) {
					t.Errorf("request IDs = %v, want %q", ids, tc.want)
				}
				if got := md.Get("other"); len(got) != 1 || got[0] != "kept" {
					t.Errorf("existing metadata = %v", got)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
