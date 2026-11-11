package domain

import (
	"fmt"

	"github.com/victorinolavida/pkg/broker"
)

type EventType string

const (
	OrderBookSnapshot EventType = "market.depth"
)

// Event represents a publishable domain event.
type Event interface {
	EventType() string
}

var _ Event = (*MarketDepthEvent)(nil)
var _ broker.Eventable = (*MarketDepthEvent)(nil)

type MarketDepthEvent struct {
	Ask        []*MarketDepthSideEvent `json:"ask"`
	Bid        []*MarketDepthSideEvent `json:"bid"`
	SecurityID string                  `json:"security_id"`
}

type MarketDepthSideEvent struct {
	Price           string `json:"price"`
	AggNominalValue string `json:"nominal_value"`
	AggQuantity     int64  `json:"quantity"`
	OrderCounts     int64  `json:"order_counts"`
	Direction       string `json:"direction"`
}

func NewMarketDepthEvent(securityID string, depth *MarketDepth) *MarketDepthEvent {
	event := &MarketDepthEvent{SecurityID: securityID}
	if depth == nil {
		return event
	}
	for _, side := range depth.Ask {
		event.Ask = append(event.Ask, &MarketDepthSideEvent{
			Price:           side.Price.String(),
			AggNominalValue: side.AggNominalValue.String(),
			AggQuantity:     side.AggQuantity.Value(),
			OrderCounts:     side.OrderCounts,
			Direction:       side.Direction.String(),
		})
	}
	for _, side := range depth.Bid {
		event.Bid = append(event.Bid, &MarketDepthSideEvent{
			Price:           side.Price.String(),
			AggNominalValue: side.AggNominalValue.String(),
			AggQuantity:     side.AggQuantity.Value(),
			OrderCounts:     side.OrderCounts,
			Direction:       side.Direction.String(),
		})
	}
	return event
}

func (m *MarketDepthEvent) EventType() string {
	return fmt.Sprintf("%s.%s", OrderBookSnapshot, m.SecurityID)
}
