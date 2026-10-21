package response

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapGRPCError translates an error returned by a downstream gRPC service into
// the HTTP status code and the message that the client receives.
//
// It is the anti-corruption boundary of the gateway: codes coming from other
// services never leak to the API consumer as gRPC semantics.
func MapGRPCError(err error) (int, string) {
	if err == nil {
		return http.StatusInternalServerError, "something went wrong"
	}

	if errors.Is(err, context.Canceled) {
		return http.StatusRequestTimeout, "request canceled"
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout, "upstream service timeout"
	}

	st, ok := status.FromError(err)
	if !ok {
		return http.StatusBadGateway, "upstream service error"
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, st.Message()
	case codes.Unauthenticated:
		return http.StatusUnauthorized, st.Message()
	case codes.PermissionDenied:
		return http.StatusForbidden, st.Message()
	case codes.NotFound:
		return http.StatusNotFound, st.Message()
	case codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted:
		return http.StatusConflict, st.Message()
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, st.Message()
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "upstream service unavailable"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "upstream service timeout"
	case codes.Canceled:
		return http.StatusRequestTimeout, "request canceled"
	default:
		return http.StatusInternalServerError, "something went wrong"
	}
}

// RenderGRPCError maps a gRPC error and writes it as an HTTP error response.
func RenderGRPCError(w http.ResponseWriter, err error) {
	statusCode, message := MapGRPCError(err)

	if statusCode >= http.StatusInternalServerError {
		slog.Error("upstream service error", slog.Any("error", err))
	}

	RenderError(w, statusCode, message)
}
