package dtos

import "time"

type CreateOrderRequest struct {
	TraderID   string `json:"trader_id"`
	SecurityID string `json:"security_id"`
	Direction  string `json:"direction" validate:"required,oneof=bid ask"`
	OrderType  string `json:"order_type" validate:"required,oneof=limit market"`
	Price      string `json:"price"`
	Yield      string `json:"yield"`
	Quantity   int64  `json:"quantity" validate:"required"`
}

type CancelOrderRequest struct {
	OrderGroupID string `json:"order_group_id" validate:"required"`
	TraderID     string `json:"trader_id"`
}

type GetOrderRequest struct {
	OrderID string `json:"order_id" validate:"required"`
}

type GetActiveOrdersRequest struct {
	SecurityID string `json:"security_id" validate:"required"`
	Direction  string `json:"direction" validate:"omitempty,oneof=bid ask"`
	TraderID   string
}

type OrderResponse struct {
	ID                string    `json:"id"` // public id
	SecurityID        string    `json:"security_id"`
	Direction         string    `json:"direction"`
	OrderType         string    `json:"order_type"`
	Price             string    `json:"price"`
	Yield             string    `json:"yield"`
	Quantity          int64     `json:"quantity"`
	NominalValue      string    `json:"nominal_value"`
	AllocatedQuantity int64     `json:"allocated_quantity"`
	AllocationPtc     string    `json:"allocation_ptc"`
	AllocationStatus  string    `json:"allocation_status"`
	Fx                string    `json:"fx"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateOrderResponse = OrderResponse

type CancelOrderResponse = OrderResponse

type GetActiveOrdersResponse struct {
	Orders []OrderResponse `json:"orders"`
}
