package csrf

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/victorinolavida/pkg/config"
)

const (
	csrfHeader = "X-CSRF-Token"
	// randomBytes is the size of the random part of a csrf token.
	randomBytes = 32
)

var (
	ErrNoCSRFToken       = errors.New("csrf: token missing")
	ErrInvalidCSRFToken  = errors.New("csrf: token invalid")
	ErrCSRFSignature     = fmt.Errorf("%w: bad signature", ErrInvalidCSRFToken)
	ErrCSRFTokenMismatch = fmt.Errorf("%w: header does not match cookie", ErrInvalidCSRFToken)
)

// Issue returns a new signed csrf token, made of a random value and its
// signature: base64url(random) + "." + base64url(HMAC-SHA256(secret, random)).
//
// The client is expected to send the value back in the X-CSRF-Token header on
// every state-changing request.
func Issue(conf *config.CookieConfig) (string, error) {
	random := make([]byte, randomBytes)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("csrf: generate random value: %w", err)
	}

	value := base64.RawURLEncoding.EncodeToString(random)

	mac := hmac.New(sha256.New, []byte(conf.Secret))
	mac.Write([]byte(value))

	return value + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func Verify(r *http.Request, conf *config.CookieConfig) error {
	cookie, err := r.Cookie(conf.CSRFName)
	if err != nil {
		return ErrNoCSRFToken
	}
	header := r.Header.Get(csrfHeader)
	if header == "" {
		return ErrNoCSRFToken
	}
	// 1. the value must be one we signed
	if !validSignature(cookie.Value, []byte(conf.Secret)) {
		return ErrCSRFSignature
	}
	// 2. the header must echo the cookie
	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
		return ErrCSRFTokenMismatch
	}
	return nil
}

func validSignature(token string, secret []byte) bool {
	i := strings.LastIndexByte(token, '.')
	if i < 0 {
		return false
	}
	value, sig := token[:i], token[i+1:]
	got, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return hmac.Equal(got, mac.Sum(nil))
}