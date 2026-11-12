package domainerrors

import "errors"

var (
	ErrInvalidTraderID         = errors.New("invalid trader id")
	ErrInvalidSecurityID       = errors.New("invalid security id")
	ErrInvalidPrice            = errors.New("invalid price")
	ErrInvalidYield            = errors.New("invalid yield")
	ErrInvalidQuantity         = errors.New("quantity must be greater than 0")
	ErrInvalidNominalValue     = errors.New("nominal value must be greater than 0")
	ErrPriceInMarketOrder      = errors.New("market order can not have price")
	ErrQtyTooLarge             = errors.New("quantity is not large")
	ErrYieldInMarketOrder      = errors.New("market order can not have yield")
	ErrInvalidFX               = errors.New("fx must be greater than 0")
	ErrInvalidDirection        = errors.New("invalid direction")
	ErrInvalidOrderType        = errors.New("invalid order type")
	ErrInvalidOrderStatus      = errors.New("invalid order status")
	ErrOrderNotFound           = errors.New("order not found")
	ErrTradeNotFound           = errors.New("trade not found")
	ErrOrderAlreadyCanceled    = errors.New("order group already canceled")
	ErrCannotCancelMarketOrder = errors.New("cannot cancel market order")
	ErrInvalidOrderId          = errors.New("order id is required")
	ErrInvalidExpiration       = errors.New("error invalid expiration value")
	ErrMarketOrderExpiration   = errors.New("market orders can not have expiration")

	ErrInvalidTakerID = errors.New("invalid taker id")
	ErrInvalidMakerID = errors.New("invalid maker id")
)
