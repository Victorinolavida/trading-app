package passwordreset

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLimiterAddressAndExpiration(t *testing.T) {
	limiter := NewLimiter()
	now := time.Now()
	limiter.now = func() time.Time { return now }
	for range 3 {
		require.True(t, limiter.Allow("192.0.2.1", "User@example.com"))
	}
	require.False(t, limiter.Allow("192.0.2.2", " user@example.com "))
	require.True(t, limiter.Allow("192.0.2.1", "other@example.com"))
	now = now.Add(15 * time.Minute)
	require.True(t, limiter.Allow("192.0.2.1", "user@example.com"))
}

func TestLimiterIP(t *testing.T) {
	limiter := NewLimiter()
	for i := range 10 {
		require.True(t, limiter.Allow("192.0.2.1", string(rune('a'+i))+"@example.com"))
	}
	require.False(t, limiter.Allow("192.0.2.1", "another@example.com"))
}
