package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/pkg/config"
)

func Test_sameSite(t *testing.T) {
	testCases := []struct {
		Value    string
		Expected http.SameSite
		Title    string
	}{
		{
			Value:    "strict",
			Expected: http.SameSiteStrictMode,
			Title:    "strict mode",
		},
		{
			Value:    "none",
			Expected: http.SameSiteNoneMode,
			Title:    "none mode",
		},
		{
			Value:    "lax",
			Expected: http.SameSiteLaxMode,
			Title:    "lax mode",
		},
		{
			Value:    "Strict",
			Expected: http.SameSiteStrictMode,
			Title:    "strinc mode with uppercase",
		},
		{
			Value:    "NoNe",
			Expected: http.SameSiteNoneMode,
			Title:    "none mode with uppercase",
		},
		{
			Value:    "LAX",
			Expected: http.SameSiteLaxMode,
			Title:    "lax mode with uppercase",
		},
		{
			Value:    "",
			Expected: http.SameSiteLaxMode,
			Title:    "not mode at all",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Title, func(t *testing.T) {
			mode := sameSite(tt.Value)
			assert.Equal(t, tt.Expected, mode)
		})
	}
}

func Test_Clear(t *testing.T) {
	conf := &config.CookieConfig{
		Name:     "test-cookie",
		CSRFName: "test-csrf",
		Path:     "/",
		Domain:   "test-domain",
		TTL:      10 * time.Minute,
		Secret:   "super_secure_cookie_secret_for_test",
		SameSite: "none",
		Secure:   true,
	}

	recorder := httptest.NewRecorder()

	// Setup: set the session and csrf cookies first
	http.SetCookie(recorder, &http.Cookie{
		Name:     conf.Name,
		Path:     conf.Path,
		Domain:   conf.Domain,
		Value:    conf.Secret + "_cookie",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   int(conf.TTL.Seconds()),
		SameSite: sameSite(conf.SameSite),
	})
	http.SetCookie(recorder, &http.Cookie{
		Name:     conf.CSRFName,
		Path:     conf.Path,
		Domain:   conf.Domain,
		Value:    conf.Secret + "_csrf",
		Secure:   true,
		MaxAge:   int(conf.TTL.Seconds()),
		SameSite: sameSite(conf.SameSite),
	})

	// Sanity check using the live header map (does NOT snapshot the recorder)
	require.Len(t, recorder.Header().Values("Set-Cookie"), 2)

	// Act
	Clear(recorder, conf)

	// Assert
	res := recorder.Result()
	defer res.Body.Close()

	cookies := res.Cookies()
	require.Len(t, cookies, 4) // 2 from setup + 2 from Clear

	// Emulate a browser: the last Set-Cookie per name wins
	last := map[string]*http.Cookie{}
	for _, c := range cookies {
		last[c.Name] = c
	}

	for _, name := range []string{conf.Name, conf.CSRFName} {
		c, ok := last[name]
		require.True(t, ok, "cookie %q not found", name)

		require.Empty(t, c.Value, "cookie %q should have empty value", name)
		require.Less(t, c.MaxAge, 0, "cookie %q should be expired", name)
		require.True(t, c.Expires.Before(time.Now()), "cookie %q Expires should be in the past", name)
		require.Equal(t, conf.Path, c.Path)
		require.Equal(t, conf.Domain, c.Domain)
		require.True(t, c.Secure, "cookie %q must be Secure (required with SameSite=None)", name)
		require.Equal(t, http.SameSiteNoneMode, c.SameSite)
	}

	// HttpOnly should match how each cookie was originally set
	require.True(t, last[conf.Name].HttpOnly)
	require.False(t, last[conf.CSRFName].HttpOnly)
}

func Test_Set(t *testing.T) {
	conf := &config.CookieConfig{
		Name:     "test-cookie",
		CSRFName: "test-csrf",
		Path:     "/",
		Domain:   "test-domain",
		TTL:      10 * time.Minute,
		Secret:   "super_secure_cookie_secret_for_test",
		SameSite: "none",
		Secure:   true,
	}
	recorder := httptest.NewRecorder()
	token := "super_secure_cookie_secret_for_test"

	err := Set(recorder, conf, token)
	require.NoError(t, err)

	res := recorder.Result()
	defer res.Body.Close()

	cookies := res.Cookies()
	require.Len(t, cookies, 2)
	for _, cookie := range res.Cookies() {
		for _, name := range []string{conf.Name, conf.CSRFName} {
			if name == cookie.Name {
				t.Logf("%v", cookie)
			}
			require.NotEmpty(t, cookie.Value)
			require.Greater(t, cookie.MaxAge, 0)
			require.Equal(t, cookie.MaxAge, int(conf.TTL.Seconds()))
			require.True(t, cookie.Expires.After(time.Now()))
			require.Equal(t, conf.Path, cookie.Path)
			require.Equal(t, conf.Domain, cookie.Domain)
			require.True(t, cookie.Secure, "cookie %q must be Secure (required with SameSite=None)", name)
			require.Equal(t, http.SameSiteNoneMode, cookie.SameSite)
		}
	}
}
