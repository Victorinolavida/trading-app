package domainerrors

import "errors"

var (
	ErrInvalidTraderID           = errors.New("invalid trader id")
	ErrInvalidSecurityID         = errors.New("invalid security id")
	ErrInvalidPrice              = errors.New("invalid price")
	ErrInvalidYield              = errors.New("invalid yield")
	ErrInvalidQuantity           = errors.New("quantity must be greater than 0")
	ErrInvalidNominalValue       = errors.New("nominal value must be greater than 0")
	ErrPriceInMarketOrder        = errors.New("market order can not have price")
	ErrYieldInMarketOrder        = errors.New("market order can not have yield")
	ErrInvalidFX                 = errors.New("fx must be greater than 0")
	ErrInvalidDirection          = errors.New("invalid direction")
	ErrInvalidOrderType          = errors.New("invalid order type")
	ErrInvalidOrderStatus        = errors.New("invalid order status")
	ErrInvalidOrderGroupStatus   = errors.New("invalid order group status")
	ErrOrderNotFound             = errors.New("order not found")
	ErrOrderGroupNotFound        = errors.New("order group not found")
	ErrOrderGroupAlreadyCanceled = errors.New("order group already canceled")
	ErrOrderAlreadyAllocated     = errors.New("order already allocated")
	ErrCannotCancelMarketOrder   = errors.New("cannot cancel market order")
)
