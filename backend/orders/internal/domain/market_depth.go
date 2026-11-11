package domain

type MarketDepthSide struct {
	Price           Decimal
	AggNominalValue Decimal
	AggQuantity     Quantity
	OrderCounts     int64
	Direction       Direction
}

type MarketDepth struct {
	Ask []*MarketDepthSide
	Bid []*MarketDepthSide
}

func ReconstructMarketDepth(
	price, nominalValue Decimal,
	quantity Quantity,
	direction Direction,
	orderCount int64,
) *MarketDepthSide {
	return &MarketDepthSide{
		Price:           price,
		AggNominalValue: nominalValue,
		AggQuantity:     quantity,
		OrderCounts:     orderCount,
		Direction:       direction,
	}
}
