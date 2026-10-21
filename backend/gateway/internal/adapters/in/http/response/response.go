package response

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const (
	authorizationHeader = "Authorization"
	maxLimitBodyRequest = 1 << 20
)

func WriteJsonResponse(w http.ResponseWriter, data any, status int) error {
	d, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(d)
	return nil
}

func RenderJson(w http.ResponseWriter, data any, status int) {
	err := WriteJsonResponse(w, data, status)
	if err != nil {
		RenderServerError(w, err)
	}
}

func RenderError(w http.ResponseWriter, status int, message any) {
	env := Envelop("error", message)
	err := WriteJsonResponse(w, env, status)
	if err != nil {
		slog.Error("error writing json response", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func RenderServerError(w http.ResponseWriter, err error) {
	slog.Error("sending error response", slog.Any("error", err))
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
	RenderError(w, http.StatusBadRequest, message)
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

func GetTokenFromRequest(r *http.Request) (string, error) {
	bearerToken := r.Header.Get(authorizationHeader)

	if bearerToken == "" {
		return "", errors.New("unauthorized")
	}

	return getCleanToken(bearerToken)
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
		slog.Error("error decodeing", slog.Any("error", err))
		return *newStruct, errors.New("invalid data")
	}
	return *newStruct, nil
}
