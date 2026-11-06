package response

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

type TokenSource string

const (
	authorizationHeader = "Authorization"
	maxLimitBodyRequest = 1 << 20

	SourceBearer  TokenSource = "bearer"
	SourceCookie  TokenSource = "cookie"
	SourceInvalid TokenSource = "invalid"
)

// ErrNoCredential is returned when the request carries neither an
// Authorization header nor a session cookie.
var ErrNoCredential = errors.New("no credentials in request")

func WriteJsonResponse(w http.ResponseWriter, data any, status int) error {
	d, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(d)
	return err
}

func RenderJson(w http.ResponseWriter, data any, status int) {
	err := WriteJsonResponse(w, data, status)
	if err != nil {
		RenderServerError(w, err)
	}
}

func RenderError(w http.ResponseWriter, status int, message any) {
	env := Envelop("error", message)
	if rw, ok := w.(*RequestWriter); ok {
		if rid := middleware.GetReqID(rw.req.Context()); rid != "" {
			env = map[string]any{
				"error":      message,
				"request_id": rid,
			}
		}
	}
	err := WriteJsonResponse(w, env, status)
	if err != nil {
		slog.Error("error writing json response", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func RenderServerError(w http.ResponseWriter, err error) {
	args := []any{slog.Any("error", err)}
	if r := RequestFromContext(w); r != nil {
		if reqID := middleware.GetReqID(r.Context()); reqID != "" {
			args = append(args, slog.String("request_id", reqID))
		}
	}
	slog.Error("sending error response", args...)
	message := "something went wrong"
	RenderError(w, http.StatusInternalServerError, message)
}

func RenderNotFound(w http.ResponseWriter) {
	message := "page not found"
	RenderError(w, http.StatusNotFound, message)
}

func RenderForbidden(w http.ResponseWriter) {
	message := "forbidden"
	RenderError(w, http.StatusForbidden, message)
}

func RenderBadRequest(w http.ResponseWriter, msg ...string) {
	message := "bad request"
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	RenderError(w, http.StatusBadRequest, message)
}

func RenderMethodNotAllowed(w http.ResponseWriter) {
	message := "method not allowed"
	RenderError(w, http.StatusMethodNotAllowed, message)
}

func RenderFatalError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		RenderServerError(w, err)
	} else {
		RenderError(w, http.StatusBadRequest, err.Error())
	}
}

func RenderUnauthorized(w http.ResponseWriter) {
	message := "unauthorized"
	RenderError(w, http.StatusUnauthorized, message)
}

func Envelop(key string, data any) map[string]any {
	return map[string]any{key: data}
}

func GetTokenFromRequest(r *http.Request, cookieName string) (string, TokenSource, error) {
	bearerToken := r.Header.Get(authorizationHeader)
	if bearerToken != "" {
		if len(bearerToken) > 7 && strings.EqualFold(bearerToken[:7], "bearer ") {
			return strings.TrimSpace(bearerToken[7:]), SourceBearer, nil
		}
		return "", SourceInvalid, ErrNoCredential
	}
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", SourceInvalid, ErrNoCredential
	}
	return strings.TrimSpace(cookie.Value), SourceCookie, nil
}

func getCleanToken(authHeader string) (string, error) {
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "bearer ") {
		return authHeader[7:], nil
	}
	return "", errors.New("invalid token")
}

func ParseBodyFromRequest[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	// 1. Limit size to avoid abuse
	r.Body = http.MaxBytesReader(w, r.Body, maxLimitBodyRequest) // 1 MiB

	newStruct := new(T)

	// 2. Decode
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&newStruct); err != nil {
		args := []any{slog.Any("error", err)}
		if reqID := middleware.GetReqID(r.Context()); reqID != "" {
			args = append(args, slog.String("request_id", reqID))
		}
		slog.Error("error decoding", args...)
		return *newStruct, errors.New("invalid data")
	}
	return *newStruct, nil
}
