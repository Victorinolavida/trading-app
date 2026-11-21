package mock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/auth/internal/application/ports"
)

func TestSender(t *testing.T) {
	sender := &Sender{}
	message := ports.Email{To: "user@example.com", Subject: "Reset password", Text: "Reset link"}
	require.NoError(t, sender.Send(context.Background(), message))

	messages := sender.Messages()
	require.Equal(t, []ports.Email{message}, messages)
	messages[0].Text = "changed"
	require.Equal(t, message, sender.Messages()[0])

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, sender.Send(ctx, message), context.Canceled)
	require.Len(t, sender.Messages(), 1)
}
