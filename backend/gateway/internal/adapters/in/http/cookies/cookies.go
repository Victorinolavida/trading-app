// Package cookies writes the browser session cookies of the gateway.
//
// Two cookies are involved: the session cookie, which carries the access token
// and is only readable by the server, and the csrf cookie, which the frontend
// reads in order to echo it back in the X-CSRF-Token header.
package cookies

import (
	"net/http"
	"strings"
	"time"

	"github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/csrf"
	"github.com/victorinolavida/pkg/config"
)

// Set writes the session cookie and the csrf cookie.
func Set(w http.ResponseWriter, conf *config.CookieConfig, token string) error {
	csrfToken, err := csrf.Issue(conf)
	if err != nil {
		return err
	}

	maxAge := int(conf.TTL.Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     conf.Name,
		Value:    token,
		Path:     conf.Path,
		Domain:   conf.Domain,
		Secure:   conf.Secure,
		HttpOnly: true,
		SameSite: sameSite(conf.SameSite),
		MaxAge:   maxAge,
	})

	// Not HttpOnly on purpose: the frontend has to read it to send it back in
	// the X-CSRF-Token header.
	http.SetCookie(w, &http.Cookie{
		Name:     conf.CSRFName,
		Value:    csrfToken,
		Path:     conf.Path,
		Domain:   conf.Domain,
		Secure:   conf.Secure,
		HttpOnly: false,
		SameSite: sameSite(conf.SameSite),
		MaxAge:   maxAge,
	})

	return nil
}

// Clear expires the session cookie and the csrf cookie.
func Clear(w http.ResponseWriter, conf *config.CookieConfig) {
	for _, name := range []string{conf.Name, conf.CSRFName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     conf.Path,
			Domain:   conf.Domain,
			Secure:   conf.Secure,
			HttpOnly: name == conf.Name,
			SameSite: sameSite(conf.SameSite),
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
		})
	}
}

func sameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
