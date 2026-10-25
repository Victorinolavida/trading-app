package dtos


type CreateOrderRequest struct {
	TraderID   string `json:"trader_id"`
	SecurityID string `json:"security_id"`
	Direction  string `json:"direction" validate:"required,oneof=bid ask"`
	OrderType  string `json:"order_type" validate:"required,oneof=limit market"`
	Price      string `json:"price"`
	Yield      string `json:"yield"`
	Quantity   string `json:"quantity" validate:"required"`
}

type CreateOrderResponse struct {
	OrderGroupID string `json:"order_group_id"`
	OrderID      string `json:"order_id"`
	Status       string `json:"status"`
}

type CancelOrderRequest struct {
	OrderGroupID string `json:"order_group_id" validate:"required"`
	TraderID     string `json:"trader_id"`
}

type CancelOrderResponse struct {
	OrderGroupID string `json:"order_group_id"`
	Status       string `json:"status"`
}

type Order struct {
	ID           string `json:"id"`
	OrderGroupID string `json:"order_group_id"`
	TraderID     string `json:"trader_id"`
	SecurityID   string `json:"security_id"`
	Direction    string `json:"direction"`
	OrderType    string `json:"order_type"`
	Price        string `json:"price"`
	Yield        string `json:"yield"`
	Quantity     string `json:"quantity"`
	NominalValue string `json:"nominal_value"`
	Fx           string `json:"fx"`
	Status       string `json:"status"`
}

type OrderGroupDetail struct {
	ID               string  `json:"id"`
	PublicID         string  `json:"public_id"`
	TraderID         string  `json:"trader_id"`
	SecurityID       string  `json:"security_id"`
	Direction        string  `json:"direction"`
	OrderType        string  `json:"order_type"`
	AllocationPC     string  `json:"allocation_pc"`
	AllocationStatus string  `json:"allocation_status"`
	OrderGroupStatus string  `json:"order_group_status"`
	WeightedPrice    string  `json:"weighted_price"`
	WeightedYield    string  `json:"weighted_yield"`
	WeightedFX       string  `json:"weighted_fx"`
	Orders           []Order `json:"orders"`
}
