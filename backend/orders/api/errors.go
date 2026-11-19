package api

import "github.com/victorinolavida/orders/internal/domainerrors"

// Domain errors re-exported so callers can classify module failures without
// importing the internal domain error package.
var (
	ErrInvalidTraderID         = domainerrors.ErrInvalidTraderID
	ErrInvalidSecurityID       = domainerrors.ErrInvalidSecurityID
	ErrInvalidPrice            = domainerrors.ErrInvalidPrice
	ErrInvalidYield            = domainerrors.ErrInvalidYield
	ErrInvalidQuantity         = domainerrors.ErrInvalidQuantity
	ErrInvalidNominalValue     = domainerrors.ErrInvalidNominalValue
	ErrPriceInMarketOrder      = domainerrors.ErrPriceInMarketOrder
	ErrQtyTooLarge             = domainerrors.ErrQtyTooLarge
	ErrYieldInMarketOrder      = domainerrors.ErrYieldInMarketOrder
	ErrInvalidFX               = domainerrors.ErrInvalidFX
	ErrInvalidDirection        = domainerrors.ErrInvalidDirection
	ErrInvalidOrderType        = domainerrors.ErrInvalidOrderType
	ErrInvalidOrderStatus      = domainerrors.ErrInvalidOrderStatus
	ErrOrderNotFound           = domainerrors.ErrOrderNotFound
	ErrTradeNotFound           = domainerrors.ErrTradeNotFound
	ErrOrderAlreadyCanceled    = domainerrors.ErrOrderAlreadyCanceled
	ErrCannotCancelMarketOrder = domainerrors.ErrCannotCancelMarketOrder
	ErrInvalidOrderId          = domainerrors.ErrInvalidOrderId
	ErrInvalidExpiration       = domainerrors.ErrInvalidExpiration
	ErrMarketOrderExpiration   = domainerrors.ErrMarketOrderExpiration
	ErrInvalidTakerID          = domainerrors.ErrInvalidTakerID
	ErrInvalidMakerID          = domainerrors.ErrInvalidMakerID
)
