package matchingengine

import (
	"fmt"

	"github.com/victorinolavida/orders/internal/domain"
)

type Rejection uint8

const (
	MatchOK Rejection = iota
	RejectTakerNoRemaining
	RejectMakerNoRemaining
	RejectTakerNotActive
	RejectMakerNotActive
	RejectSameDirection
	RejectPriceDoesNotCross
	RejectInternalCrossing
)

func canExecute(taker, maker *domain.Order) Rejection {
	switch {
	case taker.Remaining() == 0:
		return RejectTakerNoRemaining
	case maker.Remaining() == 0:
		return RejectMakerNoRemaining
	case !taker.IsActive():
		return RejectTakerNotActive
	case !maker.IsActive():
		return RejectMakerNotActive
	case taker.Direction() == maker.Direction():
		return RejectSameDirection
	case !pricesCross(taker, maker):
		return RejectPriceDoesNotCross
	}
	return MatchOK
}

func pricesCross(taker, maker *domain.Order) bool {
	if taker.IsMarket() && maker.IsMarket() {
		return false
	}
	if taker.IsMarket() || maker.IsMarket() {
		return true
	}
	takerPrice := taker.Price().Value()
	makerPrice := maker.Price().Value()
	if taker.Direction() == domain.DirectionBid {
		return takerPrice >= makerPrice
	}
	return takerPrice <= makerPrice
}

func (r Rejection) String() string {
	switch r {
	case MatchOK:
		return "ok"
	case RejectTakerNoRemaining:
		return "taker has no remaining quantity"
	case RejectMakerNoRemaining:
		return "maker has no remaining quantity"
	case RejectTakerNotActive:
		return "taker is not active"
	case RejectMakerNotActive:
		return "maker is not active"
	case RejectSameDirection:
		return "orders have the same direction"
	case RejectPriceDoesNotCross:
		return "prices do not cross"
	case RejectInternalCrossing:
		return "internal crossing"
	default:
		return fmt.Sprintf("unknown rejection (%d)", uint8(r))
	}
}
