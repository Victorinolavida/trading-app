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
	}

	recorder := httptest.NewRecorder()

	http.SetCookie(recorder, &http.Cookie{
		Name:     conf.Name,
		Path:     conf.Path,
		Domain:   conf.Domain,
		Value:    conf.Secret + "_cookie",
		Secure:   true,
		MaxAge:   int(conf.TTL.Seconds()),
		SameSite: sameSite(conf.SameSite),
	})

	http.SetCookie(recorder, &http.Cookie{
		Path:     conf.CSRFName,
		Domain:   conf.Domain,
		Value:    conf.Secret + "_csrf",
		Secure:   true,
		MaxAge:   int(conf.TTL.Seconds()),
		SameSite: sameSite(conf.SameSite),
	})
	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 2)
	t.Logf("%v\n", cookies)
	// cookie, err := request.Cookie("test")
	// assert.NoError(err)
}
