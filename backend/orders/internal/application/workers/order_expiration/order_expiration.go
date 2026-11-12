package orderexpiration

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
)

type OrderWorker struct {
	repo      ports.OrderRepository
	interval  time.Duration
	txManager ports.TxManager
	publisher ports.EventPublisher
}

func NewOrderWorker(
	repo ports.OrderRepository, txManager ports.TxManager,
	publisher ports.EventPublisher, interval time.Duration,
) *OrderWorker {
	return &OrderWorker{
		repo:      repo,
		interval:  interval,
		txManager: txManager,
		publisher: publisher,
	}
}

func (w *OrderWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("Order expiration worker started")

	// Execute immediately on startup
	err := w.execute(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Order expiration worker stopped")
			return ctx.Err()

		case <-ticker.C:
			if err := w.execute(ctx); err != nil {
				slog.Error("order expiration run failed", slog.Any("error", err))
			}
		}
	}
}

func (w *OrderWorker) execute(ctx context.Context) error {
	return w.txManager.WithTx(ctx, func(ctx context.Context) error {
		rows, err := w.repo.ExpireOrders(ctx)
		if err != nil {
			slog.Error("Error expiring orders", slog.Any("error", err))
			return err
		}

		if len(rows) == 0 {
			slog.Info("no orders to expire")
			return nil
		}

		slog.Debug("orders to expire", slog.Int("number of orders", len(rows)))
		slog.Debug("orders to expire", slog.Any("number of orders", rows))
		securities := make(map[uuid.UUID]bool)
		for _, row := range rows {
			_, ok := securities[row.SecurityID()]
			if !ok {
				securities[row.SecurityID()] = true
			}
		}

		slog.Debug("market event for securities", slog.Any("ids", securities))

		for secId := range securities {
			depth, err := w.repo.GetMarketDepth(ctx, secId)
			if err != nil {
				slog.Error("error getting depth", slog.Any("error", err))
				return err
			}
			event := domain.NewMarketDepthEvent(secId.String(), depth)

			if pubErr := w.publisher.Publish(ctx, event); pubErr != nil {
				slog.Warn("failed to publish market depth event", slog.Any("error", pubErr))
			}
		}
		return nil
	})
}
