package orderexpiration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/orders/internal/application/mocks"
	"github.com/victorinolavida/orders/internal/domain"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type stubTxManager struct{}

func (stubTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (stubTxManager) Lock(context.Context, string) error { return nil }

type stubPublisher struct{}

func (stubPublisher) Publish(context.Context, ...domain.Event) error { return nil }

func TestOrderWorker_StopsWhenContextIsCanceled(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	repo.On("ExpireOrders", mock.Anything).Return([]*domain.Order{}, nil)

	worker := NewOrderWorker(repo, stubTxManager{}, stubPublisher{}, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- worker.Start(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after the context was canceled")
	}
}
